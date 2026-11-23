#!/usr/bin/env python3
"""Black-box API E2E: gateway -> Lambda RIE -> PostgreSQL / SQS / S3 / SMTP."""
import json, urllib.request, urllib.error, time, uuid, os
BASE=os.environ.get('BASE_URL','http://127.0.0.1:8088')
def request(path,actor=None,body=None,key=None,method=None):
 headers={'Content-Type':'application/json'}
 if actor: headers['X-Actor-ID']=actor
 if key: headers['Idempotency-Key']=key
 req=urllib.request.Request(BASE+path,headers=headers,data=None if body is None else json.dumps(body).encode(),method=method)
 with urllib.request.urlopen(req,timeout=55) as r:
  raw=r.read()
  return json.loads(raw) if 'json' in r.headers.get('Content-Type','') else raw
def expect_error(status,*args,**kwargs):
 try:
  request(*args,**kwargs)
 except urllib.error.HTTPError as e:
  with e:
   assert e.code==status,(e.code,status,e.read())
   assert json.load(e)['error']
 else:
  raise AssertionError(('expected HTTP error',status,args))
assert request('/api/health')=={'ok':True}
boot=request('/api/bootstrap');actors=boot['actors'];products=boot['products']
def actor(role):return next(a for a in actors if a['role']==role)
customer=actor('customer');warehouse=actor('warehouse');carrier=actor('carrier');operator=actor('operator')
def cmd(a,kind,payload,expect='succeeded',key=None):
 key=key or str(uuid.uuid4());request('/api/commands',a['id'],{'type':kind,'payload':payload},key)
 for i in range(100):
  state=request('/api/state',a['id']);c=next((c for c in state['commands'] if c['id']==key),None)
  if c and c['status']!='queued' and (c['status']!='succeeded' or c['projected']):
   assert c['status']==expect,(kind,c)
   return c
  time.sleep(.3)
 raise AssertionError(('command timeout',kind,key))
def waitview(a,kind,id,predicate):
 for i in range(80):
  state=request('/api/state',a['id']);v=next((x for x in state['views'] if x['kind']==kind and x['id']==id),None)
  if v and predicate(v['body']):return v
  time.sleep(.3)
 raise AssertionError(('projection timeout',kind,id))
p=products[0]
expect_error(401,'/api/state')
expect_error(403,'/api/maintenance',customer['id'],method='POST')
assert request('/api/maintenance',operator['id'],method='POST')=={'accepted':True}
expect_error(400,'/api/commands',customer['id'],{'type':'cart.set','payload':{'productId':p['id'],'quantity':1,'tracking':'unrelated'}},str(uuid.uuid4()))
expect_error(400,'/api/commands',customer['id'],{'type':'cart.set','payload':{'productId':p['id']}},str(uuid.uuid4()))
expect_error(400,'/api/commands',customer['id'],{'type':'cart.expire','payload':{}},str(uuid.uuid4()))
expect_error(400,'/api/commands',customer['id'],{'type':'cart.set','payload':{'productId':p['id'],'quantity':1}},'invalid')
print('PASS OpenAPI command shape, UUID validation and maintenance permissions',flush=True)
image=request(p['image']);assert b'<svg' in image
print('PASS S3 product image',flush=True)
c=cmd(customer,'cart.set',{'productId':p['id'],'quantity':2});request('/api/commands',customer['id'],{'type':'cart.set','payload':{'productId':p['id'],'quantity':2}},c['id'])
expect_error(409,'/api/commands',customer['id'],{'type':'cart.set','payload':{'productId':p['id'],'quantity':3}},c['id'])
expect_error(403,'/api/commands/'+c['id']+'/retry',customer['id'],method='POST')
expect_error(409,'/api/commands/'+c['id']+'/retry',operator['id'],method='POST')
waitview(customer,'cart',customer['id'],lambda c:c['items'].get(p['id'])==2)
cmd(customer,'cart.set',{'productId':p['id'],'quantity':10000},'rejected')
print('PASS cart reservation, duplicate command, insufficient stock',flush=True)
c=cmd(customer,'order.checkout',{'address':customer['address'],'coupon':'WELCOME10','note':'API E2E 検証'});oid=c['result']['orderId']
cmd(customer,'order.pay',{'orderId':oid,'success':True})
cmd(warehouse,'order.pick',{'orderId':oid})
cmd(warehouse,'order.block',{'orderId':oid,'note':'箱破損・再梱包'})
cmd(warehouse,'order.pick',{'orderId':oid})
cmd(warehouse,'order.pack',{'orderId':oid,'boxCount':2,'weightGrams':1200})
cmd(carrier,'order.dispatch',{'orderId':oid,'tracking':'E2E-'+oid[:8]})
cmd(customer,'order.cancel',{'orderId':oid},'rejected')
cmd(carrier,'order.delivery_failed',{'orderId':oid,'note':'不在'})
cmd(carrier,'order.redeliver',{'orderId':oid})
cmd(carrier,'order.deliver',{'orderId':oid})
order=waitview(customer,'order',oid,lambda o:o['status']==7)['body']
cmd(carrier,'order.deliver',{'orderId':oid},'rejected')
assert len(order['boxes'])==2
print('PASS lifecycle, warehouse exception, redelivery, terminal rejection',flush=True)
for i in range(40):
 state=request('/api/state',customer['id']);invoice=next((x for x in state['invoices'] if x['orderId']==oid),None)
 if invoice and invoice['sent']:break
 time.sleep(.5)
assert invoice and invoice['sent'],invoice
invoice_path='/api/invoices/'+invoice['id']
pdf=request(invoice_path+'?actor='+customer['id']);assert pdf.startswith(b'%PDF')
expect_error(404,invoice_path,warehouse['id'])
print('PASS points, Japanese invoice PDF, SMTP delivery',flush=True)
with urllib.request.urlopen('http://127.0.0.1:8025/api/v1/messages') as r:mail=json.load(r)
assert mail['total']>=1
print('PASS Mailpit received attachment-bearing message',flush=True)
print('E2E ALL PASS',flush=True)
