import json,subprocess,sys,re
from pathlib import Path
root=Path.cwd()
def run(directory,output):
 with open(output,'w') as f:p=subprocess.run(['flutter','test','--machine'],cwd=directory,stdout=f,stderr=subprocess.STDOUT)
 tests={};fail=set();count=0
 for line in Path(output).read_text().splitlines():
  try:x=json.loads(line)
  except ValueError:continue
  if not isinstance(x,dict):continue
  if x.get('type')=='testStart':tests[x['test']['id']]=re.sub(r'#[0-9a-f]+(?=\()', '#OBJECT', x['test']['name'])
  if x.get('type')=='testDone':
   count+=1
   if x.get('result') in ['failure','error']:fail.add(tests.get(x['testID'],str(x['testID'])))
 return p.returncode,fail,count
code,fail,count=run(root,'mod-tests.jsonl')
print('Mod test events:',count,'failures:',sorted(fail))
if code==0:sys.exit(0)
if not fail:sys.exit(code)
subprocess.run(['git','fetch','--depth','1','origin','refs/tags/v1.19.6'],check=True)
base=root.parent/'bettbox-reference'
subprocess.run(['git','worktree','add','--detach',str(base),'FETCH_HEAD'],check=True)
subprocess.run(['flutter','pub','get'],cwd=base,check=True)
subprocess.run(['dart','run','build_runner','build'],cwd=base,check=True)
bcode,bfail,bcount=run(base,root/'reference-tests.jsonl')
print('Reference failures:',sorted(bfail))
new=fail-bfail
print('New regressions:',sorted(new))
if new or (bcode!=0 and not bfail):sys.exit(1)
print('PASS: only independently reproduced upstream test failures remain; see uploaded reports')
