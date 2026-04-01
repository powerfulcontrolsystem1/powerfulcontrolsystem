import sqlite3, json, urllib.request, urllib.error, hmac, hashlib, sys, time, uuid

DB_PATH = r'd:\\powerfulcontrolsystem\\backend\\db\\superadministrador.db'

# Read access token and public key from DB
conn = sqlite3.connect(DB_PATH)
cur = conn.cursor()
cur.execute("SELECT value, encrypted FROM configuraciones WHERE config_key = 'mercadopago.access_token' LIMIT 1")
row_at = cur.fetchone()
if not row_at:
    print('MISSING_CONFIG mercadopago.access_token')
    sys.exit(2)
access_token, access_enc = row_at[0], bool(row_at[1])
cur.execute("SELECT value, encrypted FROM configuraciones WHERE config_key = 'mercadopago.public_key' LIMIT 1")
row_pk = cur.fetchone()
if not row_pk:
    print('MISSING_CONFIG mercadopago.public_key')
    sys.exit(3)
public_key, public_enc = row_pk[0], bool(row_pk[1])
print('ACCESS_TOKEN_ENCRYPTED:', access_enc)
print('PUBLIC_KEY_ENCRYPTED:', public_enc)
if access_enc or public_enc:
    print('ERROR: access_token or public_key is encrypted; aborting (no key to decrypt).')
    sys.exit(4)

# Create card token (sandbox test card)
card_payload = {
    "card_number": "4509953566233704",
    "expiration_month": 12,
    "expiration_year": 2028,
    "security_code": "123",
    "cardholder": {"name": "Test User", "identification": {"type": "CC", "number": "12345678"}},
    "public_key": public_key
}
try:
    req = urllib.request.Request('https://api.mercadopago.com/v1/card_tokens', data=json.dumps(card_payload).encode('utf-8'), headers={'Content-Type':'application/json', 'Authorization': 'Bearer ' + access_token}, method='POST')
    with urllib.request.urlopen(req, timeout=30) as resp:
        resp_b = resp.read().decode('utf-8')
        j = json.loads(resp_b)
        token_id = j.get('id')
        print('CARD_TOKEN_CREATED:', json.dumps({'id': token_id, 'last_four_digits': j.get('last_four_digits')}, ensure_ascii=False))
except urllib.error.HTTPError as he:
    print('CARD_TOKEN_HTTP_ERROR', he.code, he.read().decode('utf-8'))
    sys.exit(5)
except Exception as e:
    print('CARD_TOKEN_ERROR', str(e))
    sys.exit(6)

if not token_id:
    print('NO_CARD_TOKEN')
    sys.exit(7)

# Create payment using the card token
payment_payload = {
    "transaction_amount": 70000,
    "token": token_id,
    "description": "Licencia por 30 días",
    "installments": 1,
    "payment_method_id": "visa",
    "payer": {"email": "test@example.com", "identification": {"type": "CC", "number": "12345678"}},
    "external_reference": "licencia_1_empresa_1"
}
try:
    headers_pay = {'Content-Type':'application/json', 'Authorization': 'Bearer ' + access_token, 'X-Idempotency-Key': str(uuid.uuid4())}
    req2 = urllib.request.Request('https://api.mercadopago.com/v1/payments', data=json.dumps(payment_payload).encode('utf-8'), headers=headers_pay, method='POST')
    with urllib.request.urlopen(req2, timeout=30) as resp2:
        resp2_b = resp2.read().decode('utf-8')
        pay = json.loads(resp2_b)
        print('PAYMENT_RESPONSE:', json.dumps(pay, ensure_ascii=False))
        payment_id = pay.get('id')
        payment_status = pay.get('status')
except urllib.error.HTTPError as he:
    print('PAYMENT_HTTP_ERROR', he.code, he.read().decode('utf-8'))
    sys.exit(8)
except Exception as e:
    print('PAYMENT_ERROR', str(e))
    sys.exit(9)

# Wait briefly for webhook to arrive and server to process
if payment_id:
    time.sleep(6)
    cur.execute("SELECT id, preference_id, payment_id, status, fecha_creacion FROM pagos_mercadopago WHERE payment_id = ? LIMIT 1", (str(payment_id),))
    rowp = cur.fetchone()
    print('PAGOS_ROW_AFTER:', json.dumps(rowp, default=str, ensure_ascii=False))
else:
    print('NO_PAYMENT_ID')

conn.close()
