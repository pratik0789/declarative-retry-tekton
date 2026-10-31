"""A deterministic external credential/artifact service, independent of retryctl.
No Tekton API access. Administrative snapshot/restore is the experiment harness.
"""
import json
from http.server import BaseHTTPRequestHandler,HTTPServer
state={}
class Handler(BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_GET(self):
  data=json.dumps(state,sort_keys=True).encode();self.send_response(200);self.end_headers();self.wfile.write(data)
 def do_POST(self):
  global state
  data=json.loads(self.rfile.read(int(self.headers.get('Content-Length',0))) or b'{}')
  code=200
  if self.path=='/reset':
   state={'epoch':1,'credential':None,'artifact':None,'signature':None,'scanned':None,'deployed':None,'fail_at':data['fail_at'],'injected':False,'events':[]}
  elif self.path=='/restore':state=data
  else:
   task=self.path.rsplit('/',1)[-1]
   if task==state['fail_at'] and not state['injected']:
    state['injected']=True;code=503
   elif task=='acquire':state['credential']=state['epoch']
   elif task=='publish':
    if state['credential']!=state['epoch']:code=409
    else:state['artifact']=state['epoch']
   elif task=='sign':
    if state['credential']!=state['epoch'] or state['artifact']!=state['epoch']:code=409
    else:state['signature']=state['epoch']
   elif task=='scan':
    if state['signature']!=state['epoch']:code=409
    else:state['scanned']=state['epoch']
   elif task=='deploy':
    if state['scanned']!=state['epoch']:code=409
    else:state['deployed']=state['epoch']
   else:code=400
   state['events'].append({'task':task,'http':code,'epoch':state['epoch']})
  self.send_response(code);self.end_headers();self.wfile.write(json.dumps(state,sort_keys=True).encode())
HTTPServer(('0.0.0.0',8000),Handler).serve_forever()
