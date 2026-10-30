"""Offline transport: stage pinned public images and a bare public Git snapshot.
Run after live.py --setup; requires existing host Docker images and bare Git repo.
All writes are to the study's fresh namespace. No source/recovery PipelineRun.
"""
import hashlib,json,pathlib,subprocess,tarfile
root=pathlib.Path(__file__).resolve().parents[2];r=root/'experiments/public-oci';out=root/'results/public-oci';ns='retry-public-oci-20261030';context='kind-retry-paper'
k=['kubectl','--context',context,'-n',ns]
def apply(o):subprocess.run(k+['apply','-f','-'],input=json.dumps(o),text=True,capture_output=True,check=True)
images=json.loads((r/'upstream/runtime-images.json').read_text())
gitimage=images['ghcr.io/tektoncd-catalog/git-clone:v1.1.0']['repo_digests'][0]
apply({'apiVersion':'apps/v1','kind':'Deployment','metadata':{'name':'git-source'},'spec':{'replicas':1,'selector':{'matchLabels':{'app':'paper-git-source'}},'template':{'metadata':{'labels':{'app':'paper-git-source'}},'spec':{'securityContext':{'fsGroup':65532},'containers':[{'name':'git','image':gitimage,'command':['git','daemon','--reuseaddr','--base-path=/git','--export-all','--listen=0.0.0.0','--port=9418','/git'],'securityContext':{'runAsUser':65532,'runAsNonRoot':True},'volumeMounts':[{'name':'data','mountPath':'/git'}]}],'volumes':[{'name':'data','emptyDir':{}}]}}}})
apply({'apiVersion':'v1','kind':'Service','metadata':{'name':'git-source'},'spec':{'selector':{'app':'paper-git-source'},'ports':[{'port':9418,'targetPort':9418}]}})
subprocess.run(k+['rollout','status','deployment/git-source','--timeout=300s'],check=True,timeout=330)
repo=pathlib.Path('/tmp/tekton-public-skaffold.git');commit=subprocess.check_output(['git','-C',str(repo),'rev-parse','HEAD'],text=True).strip();assert commit=='035e3ddc3589c1c64cb69c563b178f34b1cccc5e'
archive=pathlib.Path('/tmp/tekton-public-skaffold-bare.tar')
with tarfile.open(archive,'w') as t:t.add(repo,arcname='skaffold.git')
with archive.open('rb') as f:subprocess.run(k+['exec','-i','deployment/git-source','--','tar','xf','-','-C','/git'],stdin=f,capture_output=True,check=True)
transports={'git_commit':commit,'git_uri':'git://git-source:9418/skaffold.git','git_archive_sha256':hashlib.sha256(archive.read_bytes()).hexdigest(),'images':{}}
for im,label in [('mirror.gcr.io/library/golang:1.18','golang'),('mirror.gcr.io/library/alpine:3','alpine'),('alpine:3.20','foreign')]:
 print('Staging pinned image archive',im,flush=True);path=pathlib.Path('/tmp/tekton-public-'+label+'.tar')
 subprocess.run(['docker','save','--platform','linux/amd64','-o',str(path),im],check=True,timeout=300)
 with path.open('rb') as f:subprocess.run(k+['exec','-i','deployment/oci-admin','--','/bin/sh','-c','cat > /tmp/'+label+'.tar'],stdin=f,capture_output=True,check=True,timeout=300)
 uploaded=subprocess.check_output(k+['exec','deployment/oci-admin','--','sha256sum','/tmp/'+label+'.tar'],text=True).split()[0]
 assert uploaded==hashlib.sha256(path.read_bytes()).hexdigest()
 info=json.loads(subprocess.check_output(['docker','image','inspect',im],text=True))[0]
 transports['images'][im]={'archive_sha256':hashlib.sha256(path.read_bytes()).hexdigest(),'archive_bytes':path.stat().st_size,'source_repo_digests':info['RepoDigests'],'source_rootfs':info['RootFS'],'source_architecture':info['Architecture'],'source_os':info['Os'],'source_config':info['Config'],'transport':'docker-archive:/tmp/'+label+'.tar'}
(out/'offline-transports.json').write_text(json.dumps(transports,indent=2)+'\n')
for variant in ['tagged','digest']:
 p=json.loads((r/('profile-'+variant+'.json')).read_text());p['metadata']={'name':'public-oci-'+variant}
 clone=next(t for t in p['spec']['tasks'] if t['name']=='clone');next(v for v in clone['params'] if v['name']=='url')['value']=transports['git_uri']
 build=next(t for t in p['spec']['tasks'] if t['name']=='build');build['taskSpec']['steps'][0]['command']=['/kaniko/executor']
 build['params']=[v for v in build['params'] if v['name']!='BUILDER_IMAGE']+[{'name':'BUILDER_IMAGE','value':'gcr.io/kaniko-project/executor@sha256:9e69fd4330ec887829c780f5126dd80edc663df6def362cd22e79bcdf00ac53f'}]
 extra=next(v for v in build['params'] if v['name']=='EXTRA_ARGS')['value']
 if '--ignore-path=/product_uuid' not in extra:extra.append('--ignore-path=/product_uuid')
 apply(p);actual=subprocess.check_output(k+['get','pipeline','public-oci-'+variant,'-o','json'],text=True);(r/('profile-'+variant+'.json')).write_text(actual);(out/variant/'pipeline.json').write_text(actual)
print('OFFLINE TRANSPORTS READY; no source/recovery PipelineRun executed',flush=True)
