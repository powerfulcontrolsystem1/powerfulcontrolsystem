#!/usr/bin/env python3
"""Harden a direct public host Nginx; --apply backs up and validates before reload.

Run only after checking there is no trusted CDN/LB in front of this Nginx.
The main PCS site is explicit; other sites only get spoofed-header sanitization.
"""
import argparse
import datetime
import pathlib
import re
import shutil
import subprocess

HTTP_POLICY = """# Managed PCS HTTP abuse protection. Context: http.
server_tokens off;
client_header_timeout 15s;
client_body_timeout 60s;
send_timeout 60s;
map $uri $pcs_auth_limit_key {
    default "";
    ~^/super/api/administradores/(login|register|solicitar_recuperacion|restablecer_password)/?$ $binary_remote_addr;
    ~^/api/empresa/usuarios/(login|establecer_password|solicitar_recuperacion_password|restablecer_password)/?$ $binary_remote_addr;
    ~^/auth/google/login/?$ $binary_remote_addr;
}
limit_req_zone $binary_remote_addr zone=pcs_public_ip:20m rate=60r/s;
limit_req_zone $pcs_auth_limit_key zone=pcs_auth_ip:10m rate=1r/s;
limit_conn_zone $binary_remote_addr zone=pcs_connections:10m;
map $scheme $pcs_hsts { default ""; https "max-age=31536000"; }
"""

SERVER_POLICY = """# Managed PCS limits. Context: server. Keys use the socket IP.
limit_req zone=pcs_public_ip burst=240 nodelay;
limit_req zone=pcs_auth_ip burst=20 nodelay;
limit_req_status 429;
limit_conn pcs_connections 100;
limit_conn_status 429;
proxy_hide_header Strict-Transport-Security;
add_header Strict-Transport-Security $pcs_hsts always;
add_header Content-Security-Policy "base-uri 'self'; object-src 'none'; frame-ancestors 'self'" always;
add_header X-Content-Type-Options nosniff always;
add_header Referrer-Policy strict-origin-when-cross-origin always;
"""


def site_policy(content, main=False):
    content = re.sub(r"(proxy_set_header\s+X-Forwarded-For\s+)[^;]+;",
                     r"\g<1>$remote_addr;", content)
    # Reset forwarded host even if the caller supplied one.
    content = re.sub(r"(proxy_set_header\s+X-Forwarded-Host\s+)[^;]+;",
                     r"\g<1>$host;", content)
    if "proxy_set_header X-Forwarded-Host" not in content:
        content = re.sub(r"(proxy_set_header\s+X-Forwarded-Proto\s+[^;]+;)",
                         r"\1\n        proxy_set_header X-Forwarded-Host $host;", content)
    marker = "include /etc/nginx/snippets/pcs-http-hardening.conf;"
    if main and marker not in content:
        content = re.sub(r"(^\s*server_name\s+[^;]+;)", r"\1\n    " + marker,
                         content, flags=re.MULTILINE)
    return content


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--main-site", required=True, type=pathlib.Path)
    parser.add_argument("--apply", action="store_true")
    args = parser.parse_args()
    main_site = args.main_site.resolve(strict=True)
    enabled = pathlib.Path("/etc/nginx/sites-enabled")
    sites = {p.resolve(strict=True) for p in enabled.iterdir() if p.is_file()}
    if main_site not in sites:
        raise SystemExit("Main site must be an enabled Nginx site")
    changes = {
        pathlib.Path("/etc/nginx/conf.d/pcs-http-hardening.conf"): HTTP_POLICY,
        pathlib.Path("/etc/nginx/snippets/pcs-http-hardening.conf"): SERVER_POLICY,
    }
    for site in sites:
        changes[site] = site_policy(site.read_text(), site == main_site)
    changes = {p: v for p, v in changes.items() if not p.exists() or p.read_text() != v}
    print("CHANGED_FILES=" + str(len(changes)))
    if not args.apply or not changes:
        return
    subprocess.run(["nginx", "-t"], check=True)
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    backup = pathlib.Path("/var/backups/pcs-security") / stamp
    backup.mkdir(parents=True, mode=0o700)
    existing = set()
    # Complete all backups before modifying any configuration.
    for path in changes:
        if path.exists():
            existing.add(path)
            target = backup / path.relative_to("/")
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(path, target)
    written = []
    try:
        for path, content in changes.items():
            path.parent.mkdir(parents=True, exist_ok=True)
            written.append(path)
            path.write_text(content)
        subprocess.run(["nginx", "-t"], check=True)
        subprocess.run(["systemctl", "reload", "nginx"], check=True)
    except BaseException:
        for path in written:
            if path in existing:
                shutil.copy2(backup / path.relative_to("/"), path)
            else:
                path.unlink(missing_ok=True)
        subprocess.run(["nginx", "-t"], check=True)
        subprocess.run(["systemctl", "reload", "nginx"], check=True)
        raise
    print("APPLIED_AND_RELOADED; BACKUP=" + str(backup))


if __name__ == "__main__":
    main()
