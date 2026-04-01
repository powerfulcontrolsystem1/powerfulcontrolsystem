import sqlite3, json, urllib.request, urllib.error, hmac, hashlib, sys

# Paths and DB
DB_PATH = r'd:\powerfulcontrolsystem\backend\db\superadministrador.db'
LOCAL_CREATE_PREF = 'http://localhost:8080/mercadopago/create_preference'
LOCAL_WEBHOOK = 'http://localhost:8080/mercadopago/webhook'

# Connect to DB and get webhook secret
conn = sqlite3.connect(DB_PATH)
cur = conn.cursor()
cur.execute("SELECT value, encrypted FROM configuraciones WHERE config_key = 'mercadopago.webhook_secret' LIMIT 1")
row = cur.fetchone()
if not row:
    print('NO_SECRET')
    sys.exit(2)
secret = row[0]
enc = bool(row[1])
print('SECRET_ENCRYPTED:', enc)
print('SECRET_PREVIEW:', (secret[:4] + '...' + secret[-4:]) if secret else '')

# Create a test preference via local server
req_body = json.dumps({"licencia_id": 1, "empresa_id": 1, "payer_email": "test@example.com", "payer_name": "Test User"}).encode('utf-8')
req = urllib.request.Request(LOCAL_CREATE_PREF, data=req_body, headers={'Content-Type':'application/json'}, method='POST')
try:
    with urllib.request.urlopen(req, timeout=30) as resp:
        resp_b = resp.read()
        mp = json.loads(resp_b.decode('utf-8'))
        print('PREFERENCE_RESPONSE:', json.dumps(mp, ensure_ascii=False))
except Exception as e:
    print('CREATE_PREF_ERROR:', str(e))
    sys.exit(3)

prefid = mp.get('id')
if not prefid:
    print('NO_PREF_ID')
    sys.exit(4)

# Build webhook payload referencing the created preference
payload = {"data": {"resource": {"id": "manual_test_12345", "preference_id": prefid}}}
body_raw = json.dumps(payload, separators=(',', ':'))
# Compute HMAC-SHA256 signature
sig = hmac.new(secret.encode('utf-8'), body_raw.encode('utf-8'), hashlib.sha256).hexdigest()
sig_header = 'sha256=' + sig

# Send POST to local webhook endpoint with signature header
req2 = urllib.request.Request(LOCAL_WEBHOOK, data=body_raw.encode('utf-8'), headers={'Content-Type':'application/json','X-Hub-Signature': sig_header}, method='POST')
try:
    with urllib.request.urlopen(req2, timeout=30) as r2:
        st = r2.status
        rb = r2.read().decode('utf-8')
        print('WEBHOOK_STATUS:', st)
        print('WEBHOOK_BODY:', rb)
except urllib.error.HTTPError as he:
    print('WEBHOOK_HTTP_ERROR:', he.code, he.read().decode('utf-8'))
    sys.exit(5)
except Exception as e:
    print('WEBHOOK_ERROR:', str(e))
    sys.exit(6)

# Inspect pagos_mercadopago row for the preference
cur.execute("SELECT id, preference_id, payment_id, status, fecha_creacion FROM pagos_mercadopago WHERE preference_id = ? LIMIT 1", (prefid,))
rowp = cur.fetchone()
print('PAGOS_ROW:', json.dumps(rowp, default=str, ensure_ascii=False))
conn.close()
