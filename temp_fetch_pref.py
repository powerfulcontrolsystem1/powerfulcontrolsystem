import sqlite3, json, urllib.request, sys
DB = r'd:\\powerfulcontrolsystem\\backend\\db\\superadministrador.db'
PREF = '3275096812-b993912d-b06e-4e15-8ad0-aaf431d07ede'

conn = sqlite3.connect(DB)
cur = conn.cursor()
cur.execute("SELECT value, encrypted FROM configuraciones WHERE config_key = 'mercadopago.access_token' LIMIT 1")
row = cur.fetchone()
if not row:
    print('NO_ACCESS_TOKEN')
    sys.exit(2)
access = row[0]

req = urllib.request.Request('https://api.mercadopago.com/checkout/preferences/' + PREF, headers={'Authorization': 'Bearer ' + access})
with urllib.request.urlopen(req, timeout=30) as r:
    data = r.read().decode('utf-8')
    obj = json.loads(data)
    out = {
        'id': obj.get('id'),
        'items': obj.get('items'),
        'site_id': obj.get('site_id'),
        'notification_url': obj.get('notification_url'),
        'payment_methods': obj.get('payment_methods'),
        'sandbox_init_point': obj.get('sandbox_init_point'),
    }
    print(json.dumps(out, ensure_ascii=False))

conn.close()
