"""Read-only final input-binding capture after native study completion."""
import subprocess,json,time,socket,urllib.request,hashlib,pathlib
root=pathlib.Path(__file__).resolve().parents[2];r=root/'results/public-oci';k=['kubectl','--context','kind-retry-paper','-n','retry-public-oci-20261030']
f=subprocess.Popen(k+['port-forward','service/registry-source','29175:5000'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
try:
 for _ in range(100):
  try:
   with socket.create_connection(('127.0.0.1',29175),timeout=.2):break
  except OSError:time.sleep(.2)
 d={'base_images':{}}
 def get(p):
  req=urllib.request.Request('http://127.0.0.1:29175/v2/'+p,headers={'Accept':'application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json'})
  with urllib.request.urlopen(req,timeout=30) as q:return q.read()
 for label,tag in [('golang','1.18'),('alpine','3')]:
  b=get('base/'+label+'/manifests/'+tag);m=json.loads(b);c=get('base/'+label+'/blobs/'+m['config']['digest']);d['base_images'][label]={'body_sha256':hashlib.sha256(b).hexdigest(),'config_sha256':hashlib.sha256(c).hexdigest()}
 for resource in ['pipelineruns','taskruns']:
  (r/('all-native-'+resource+'.json')).write_bytes(subprocess.check_output(k+['get',resource,'-o','json']))
 d['git_head']=subprocess.check_output(k+['exec','deployment/git-source','--','git','--git-dir=/git/skaffold.git','rev-parse','HEAD'],text=True).strip()
 d['git_pods']=json.loads(subprocess.check_output(k+['get','pods','-l','app=paper-git-source','-o','json']))
 (r/'end-input-bindings.json').write_text(json.dumps(d,indent=2)+'\n');print('Final input bindings captured')
finally:f.terminate();f.wait(timeout=10)
