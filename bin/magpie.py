#!/usr/bin/python
#-*-coding:utf-8-*- 
import argparse
import os
import sys
import urllib
import time
import ConfigParser
import logging
import logging.handlers
import json
import pycurl
import StringIO

#Logging
LOG_FILE="../log/yarn-on-docker-" + time.strftime('%Y-%m-%d',time.localtime(time.time())) + ".log"
handler = logging.handlers.RotatingFileHandler(LOG_FILE, maxBytes = 128*1024*1024, backupCount = 5)
fmt = '%(asctime)s - %(filename)s:%(lineno)s - %(name)s - %(message)s'
formatter = logging.Formatter(fmt)
handler.setFormatter(formatter)
logger = logging.getLogger('yarn_on_docker')
logger.addHandler(handler)
logger.setLevel(logging.DEBUG)

#Get config from conf.ini
config = ConfigParser.ConfigParser()
config.readfp(open("../conf/conf.ini","rb"))
cluster_name=config.get("clusters","cluster_name").split(",")
cluster_dict={}
cluster_dict=cluster_dict.fromkeys(cluster_name)
shipyard_address=config.get("shipyard","shipyard_address")
shipyard_port=config.get("shipyard","shipyard_port")
shipyard_user=config.get("shipyard","shipyard_user")
shipyard_passwd=config.get("shipyard","shipyard_passwd")
swarm_master_ip=config.get("clusters","swarm_master_ip")
swarm_master_port=config.get("clusters","swarm_master_port")

for i in range(0,len(cluster_name)):
  c=cluster_name[i]
  cluster_dict[cluster_name[i]]=config.get("resource_managers",cluster_name[i])

#Decommission the nodemanagers and delete the containers on the host.
def offline(hostname):
  for i in range(0,len(cluster_name)):
    cname=cluster_name[i]
    nms=[]
    nms_str=""
    cns=get_node_containers(hostname)
    for k in cns:
      c=cns[k].encode("utf-8")
      if c.find(cname)!=-1:
        nms.append(k+'\n')
        nms_str=nms_str+k +'\\n'
    decommising_nodes_str=nms_str[:-2]
    if decommising_nodes_str != "":
      print "DECOMISSING - CLUSTER:" + cname + " CONTAINERS:" + "".join(nms_str.replace('\\n'," "))
      logger.info("DECOMISSING - CLUSTER:" + cname + " CONTAINERS:" + "".join(nms_str.replace('\\n'," ")))
      #write nodemanagers to exclude files
      os.popen('''ssh -n root@''' + cluster_dict[cname] + ' "echo -e ' + "'" + decommising_nodes_str + ''''>>/usr/local/hadoop/etc/hadoop/test.txt"''').read()
      #os.popen('''ssh -n root@''' + cluster_dict[cname] + ' "su - hadoop -c ' + '"yarn rmadmin -refreshNodes"').read()

#Inspect the  container distribution
def inspect():
  #Show the yarn cluster status
  print "======================YARN CLSUTER STATUS==========================="
  print "CLUSTER\tTOTAL\tACTIVE\tDECOMISSIONED\tLOST\tUNHEALTHY"
  total=0
  total_active=0
  total_decommissioned=0
  total_lost=0
  total_unhealthy=0
  if args.clustername :
    d=get_yarn_status(args.clustername)
    print d["clustername"]+"\t" + str(d["num_total_nms"]) + "\t" +str(d["num_active_nms"]) +"\t"+ str(d["num_decommissioned_nms"]) +"\t"+ str(d["num_lost_nms"]) +"\t"+ str(d["num_unhealthy_nms"])
  else:
    for i in range(0,len(cluster_name)):
      d=get_yarn_status(cluster_name[i])
      total+=d["num_total_nms"] 
      total_active+=d["num_active_nms"]
      total_decommissioned+=d["num_decommissioned_nms"]
      total_lost+=d["num_lost_nms"]
      total_unhealthy+=d["num_unhealthy_nms"]
      print d["clustername"]+"\t"+str(d["num_total_nms"]) + "\t" + str(d["num_active_nms"]) +"\t"+ str(d["num_decommissioned_nms"]) +"\t"+ str(d["num_lost_nms"]) +"\t"+ str(d["num_unhealthy_nms"])
    print "--------------------------------------------------------------------"
    print "TOTAL\t"+ str(total) + "\t" + str(total_active)+ "\t" + str(total_decommissioned) + "\t" + str(total_lost) + "\t" + str(total_unhealthy) + "\n"
  #Show the docker cluster distribution
  print "======================DOCKER CLSUTER STATUS========================="
  print "CLUSTER\tTOTAL\tRUNNING\tEXITED"
  #If the clustername not specified then analysis the whole cluster
  total_num=0
  running_num=0
  exited_num=0
  if args.clustername:
    s=get_swarm_status(args.clustername) 
    print args.clustername +"\t"+ str(s["total"]) + "\t"+ str(s["running"]) + "\t"+ str(s["exited"])
  else:
    for i in range(0,len(cluster_name)):
      cname=cluster_name[i]
      status=get_swarm_status(cname)
      c_total_num=status["total"]
      c_running_num=status["running"]
      c_exited_num=status["exited"]
      total_num+=c_total_num
      running_num+=c_running_num
      exited_num+=c_exited_num
      print cname +"\t"+ str(c_total_num) + "\t"+ str(c_running_num) + "\t"+ str(c_exited_num)
    print "--------------------------------------------------------------------"
    print "TOTAL\t"+str(total_num)+"\t"+str(running_num)+"\t"+str(exited_num)+"\n"
  if args.view:
    if args.clustername:
      c=args.clustername
    else:
      c=""
    s=get_swarm_status(c)
    print "====================RUNNING DOCKERS DISTRIBUTION===================="
    d=sorted(s["running_distribution"].iteritems(), key=lambda d:d[0])
    if len(d)==0:
      print "No running contianers."
    else:
      print "HOSTNAME\tNUM"
      for i in range(len(d)):
        print d[i][0]+"\t"+str(d[i][1])
    print "====================EXITED DOCKERS DISTRIBUTION====================="
    d=sorted(s["exited_distribution"].iteritems(), key=lambda d:d[0])
    if len(d)==0:
      print "No exited containers."
    else:
      print "HOSTNAME\tNUM"
      for i in range(len(d)):
        print d[i][0]+"\t"+str(d[i][1])

#Remove a container,arg:contianer name or ID
def remove_container(container):
  url="http://"+swarm_master_ip+":"+swarm_master_port+"/containers/"+container+"?force=1"
  container_url="http://"+swarm_master_ip+":"+swarm_master_port+"/containers/" + container +"/json"
  c=pycurl.Curl()
  b= StringIO.StringIO()
  c.setopt(c.URL,container_url)
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json'])
  c.perform()
  d=json.loads(b.getvalue())
  hostname=d["Node"]["Name"]
  host_ip=d["Node"]["IP"]
  container_name=d["Name"].split("/")[1]
  container_id=d["Config"]["Hostname"]
  print "DELETE CONTAINER - " + "CONTAINER_NAME:" + container_name + " CONTAINER_ID:" + container_id + " HOSTNAME:" + hostname + " HOST_IP:" + host_ip
  c.setopt(c.URL,url)
  c.setopt(c.CUSTOMREQUEST,"DELETE") 
  print "Deleting..."
  c.perform()
  status_code=c.getinfo(pycurl.HTTP_CODE)
  logger.info("DELETE CONTAINER - " + "CONTAINER_NAME:" + container_name + " CONTAINER_ID:" + container_id + " HOSTNAME:" + hostname + " HOST_IP:" + host_ip + " STATUS_CODE:" + str(status_code))
  if status_code == 204:
    print "Delete finished."
  elif status_code == 400:
    print "ERROR:400 bad parameter!"
  elif status_code == 404:
    print "ERROR:404 no such container!"
  elif status_code == 409:
    print "ERROR:409 confilct!"
  elif status_code == 500:
    print "ERROR:500 server error!"

#Rename a container
def rename_container(oldname,newname):
  url="http://"+swarm_master_ip+":"+swarm_master_port+"/containers/"+oldname+"/rename?name="+newname
  c = pycurl.Curl()
  c.setopt(c.URL,str(url))
  c.setopt(c.CUSTOMREQUEST,"POST")
  c.perform()
  status_code=c.getinfo(pycurl.HTTP_CODE)
  logger.info("RENAME CONTAINER - " + "OLD_NAME:" + oldname + " NEW_NAME:" + newname + " STATUS_CODE:" + str(status_code))
  if status_code == 204:
    print "OK."
  elif status_code == 400:
    print "ERROR:400 bad parameter!"
  elif status_code == 409:
    print "ERROR:409 confilct!"
  elif status_code == 500:
    print "ERROR:500 server error!"

#Delete the containers on the host
def delete(hostname):
  url="http://"+swarm_master_ip+":"+swarm_master_port+"/containers/json?all=1"
  c = pycurl.Curl()
  c.setopt(c.URL, url)
  b= StringIO.StringIO()
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json'])
  c.perform()
  res=json.loads(b.getvalue())
  for i in range(len(res)):
    name=res[i]["Names"][0]
    h=name.split("/")[1]
    container_name=name.split("/")[2]
    if h==hostname:
      container_name=name.split("/")[2]
      remove_container(str(container_name))

#Get swarm cluster status
def get_swarm_status(clustername):
  result={"running":0,"exited":0,"total":0,"running_distribution":{},"exited_distribution":{}}
  url="http://"+swarm_master_ip+":"+swarm_master_port+"/containers/json?all=1"
  c = pycurl.Curl()
  b= StringIO.StringIO()
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.URL, url)
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json'])
  c.perform()
  res=json.loads(b.getvalue())
  rd={}
  ed={}
  for i in range(0,len(res)):
    network_mode=res[i]["HostConfig"]["NetworkMode"]
    state=res[i]["State"]
    name=res[i]["Names"][0]
    hostname=str(name.split("/")[1])
    container_name=str(name.split("/")[2])
    if container_name.find(clustername)!=-1:
      #print name.split("/")[1] + "\t" + name.split("/")[2] + "\t" + state
      result["total"]+=1
      if state == "running":
        result["running"]+=1
        if rd.has_key(hostname):
          rd[hostname]+=1 
        else:
          rd[hostname]=1
      if state == "exited":
        result["exited"]+=1
        if ed.has_key(hostname):
          ed[hostname]+=1
        else:
          ed[hostname]=1
  result["running_distribution"]=rd
  result["exited_distribution"]=ed
  return result 

#Get the container names of a swarm node
def get_node_containers(hostname):
  url="http://"+swarm_master_ip+":"+swarm_master_port+"/containers/json"
  c = pycurl.Curl()
  b= StringIO.StringIO()
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.URL, url)
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json'])
  c.perform()
  res=json.loads(b.getvalue())
  container_id=""
  c_hostnames=[]
  result={}
  for i in range(0,len(res)):
    name=res[i]["Names"][0]
    h=str(name.split("/")[1])
    if hostname==h:
      container_id=res[i]["Id"]
      container_name=name.split("/")[2]
      container_hostname=container_id[0:12]
      result[container_hostname]=container_name
  return result

#Get swarm nodes infomation
def get_swarm_nodes():
  token=get_auth_token()
  url=shipyard_address+":"+shipyard_port+"/api/nodes"
  c=pycurl.Curl()
  b= StringIO.StringIO()
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.URL, url)
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json','X-Access-Token:admin:'+token])
  c.perform()
  res=json.loads(b.getvalue())
  result={}
  for i in range(len(res)):
   r=res[i]
   result[r["name"]]={"addr":r["addr"],"containers":r["containers"],"reserved_cpus":r["reserved_cpus"],"labels":r["labels"]}
  return result
 
#Get yarn cluster status
def get_yarn_status(clustername):
  if not cluster_dict.has_key(clustername):
    print "The cluster does not exited!"
    quit() 
  rm_ip=cluster_dict[clustername]
  url="http://" + rm_ip + ":8088/ws/v1/cluster/metrics"
  d=json.loads(urllib.urlopen(url).read())
  num_total_nms=d["clusterMetrics"]["totalNodes"]
  num_active_nms=d["clusterMetrics"]["activeNodes"]
  num_decommissioned_nms=d["clusterMetrics"]["decommissionedNodes"]
  num_lost_nms=d["clusterMetrics"]["lostNodes"]
  num_unhealthy_nms=d["clusterMetrics"]["unhealthyNodes"]
  return {"clustername":clustername,"num_total_nms":num_total_nms,"num_active_nms":num_active_nms,"num_decommissioned_nms":num_decommissioned_nms,"num_lost_nms":num_lost_nms,"num_unhealthy_nms":num_unhealthy_nms}

#Get the contianer status
def get_container(c_name):
  url="http://"+swarm_master_ip+":"+swarm_master_port+"/containers/json"
  c = pycurl.Curl()
  b= StringIO.StringIO()
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.URL, url)
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json'])
  c.perform()
  res=json.loads(b.getvalue())
  container_id=""
  for i in range(0,len(res)):
    name=res[i]["Names"][0]
    container_name=str(name.split("/")[2])
    if container_name==c_name:
      container_id=res[i]["Id"]
      return res[i]

#Yarn cluster scaling method
def scale(clustername,num,prefix):
  c_name=config.get("base_container",clustername)
  c=get_container(c_name)
  name=c["Names"][0]
  hostname=name.split("/")[1]
  container_id=c["Id"]
  token=get_auth_token()
  url=shipyard_address+":"+shipyard_port+"/api/containers/"+str(container_id)+"/scale?n="+num+ "&nodename="+hostname
  c = pycurl.Curl()
  b= StringIO.StringIO()
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.URL,str(url))
  c.setopt(c.CUSTOMREQUEST,"POST")
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json','X-Access-Token:admin:'+token])
  print "CLUSTER:" + clustername + " BASE_CONTAINER:"+c_name + " NUM:"+num + " LOCATION:"+hostname
  c.perform()
  status_code=c.getinfo(c.HTTP_CODE)
  logger.info("SCALE - CLUSTER:" + clustername + " BASE_CONTAINER:"+c_name + " NUM:"+num + " LOCATION:"+hostname + " STATUS_CODE:"+ str(status_code))
  if status_code != 200:
    print "ERROR:"+str(status_code)
  else:
    res=json.loads(b.getvalue())
    containers=res["Scaled"]
    for i in range(len(containers)):
      logger.info("CREATE CONTAINER -" + " CONTAINER_ID:"+containers[i])
      newname=prefix + "-" + time.strftime('%Y%m%d%H%M%S',time.localtime(time.time()))+"-"+str(i)
      print "Rename container " + containers[i] + " to " + newname
      rename_container(containers[i],newname)

#Get the containers dictribution on swarm ndoes
def swarm_nodes_distribution():
  print "=======================DOCKER CONTAINERS DISTRIBUTION========================"
  print "HOSTNAME\tNUM\tIP\tRESERVED_CPUS"
  nodes=sorted(get_swarm_nodes().iteritems(), key=lambda d:d[0])
  for i in range(len(nodes)):
    print nodes[i][0]+"\t"+nodes[i][1]["containers"]+ "\t" + nodes[i][1]["addr"].split(":")[0] + "\t" +nodes[i][1]["reserved_cpus"].replace(" ","")

#Get shipyard auth_token
def get_auth_token():
  url=shipyard_address+":"+shipyard_port+"/auth/login"
  post_data=json.dumps({"username":shipyard_user,"password":shipyard_passwd})
  c = pycurl.Curl()
  b= StringIO.StringIO()
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.URL, url)
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json'])
  c.setopt(c.POST, 1)
  c.setopt(c.POSTFIELDS, post_data)
  c.perform()
  token=json.loads(b.getvalue())["auth_token"]
  return str(token)

#Get yarn cluster running contianers list
def get_yarn_running_containers(clustername):
  url="http://"+swarm_master_ip+":"+swarm_master_port+"/containers/json"
  c = pycurl.Curl()
  b= StringIO.StringIO()
  c.setopt(c.WRITEFUNCTION, b.write)
  c.setopt(c.URL, url)
  c.setopt(c.HTTPHEADER, ["Content-Type:application/json",'Accept: application/json'])
  c.perform()
  res=json.loads(b.getvalue())
  container_id=""
  container_list=[]
  for i in range(0,len(res)):
    name=res[i]["Names"][0]
    container_name=str(name.split("/")[2].split("-"[0]))
    #TODO yarn11-nm1 contain yarn1 and yarn11
    if container_name.find(clustername)!=-1:
      container_id=res[i]["Id"]
      container_list.append(res[i]["Id"][0:12])
  return container_list
 
#Get yarn cluster running nodemanagers list
def get_yarn_running_nms(rm_ip):
  url="http://" + rm_ip + ":8088/ws/v1/cluster/nodes"
  d=json.loads(urllib.urlopen(url).read())
  nodes=d["nodes"]["node"]
  node_list=[]
  for i in range(len(nodes)):
    if nodes[i]["state"]=="RUNNING":
      node_list.append(nodes[i]["id"].split(":")[0])
  return node_list

#Compare the docker containers and nodemanagers.
def compare(clustername):
  if not cluster_dict.has_key(clustername):
    print "The cluster does not exited!"
    quit()
  rm_ip=cluster_dict[clustername]
  nm_list=get_yarn_running_nms(rm_ip)
  container_list=get_yarn_running_containers(clustername)
  plus=[]
  if len(nm_list)>len(container_list):
    print "More " + str(len(nm_list)-len(container_list)) + " active nodemanager than docker containers."
    for i in range(len(nm_list)):
      if nm_list[i] not in contianer_list:
        plus.append(nm_list[i])
    for i in range(len(plus)):
      print plus[i]
  elif len(nm_list)<len(container_list):
    print "More " + str(len(container_list)-len(nm_list)) + " docker containers than active nodemanager."
    for i in range(len(container_list)):
      if container_list[i] not in nm_list:
        plus.append(container_list[i])
    for i in range(len(plus)):
      print plus[i]
  else:
    print "Docker containers are the same with yarn nodemanagers."
  
#Functon entrance
def main(args):
  if args.offline:
    if args.hostname:
      offline(args.hostname) 
    else:
      print "You must specify the hostname with -n argument."
  if args.inspect:
    inspect()
  if args.delete:
    if args.hostname:
      delete(args.hostname)
    else:
      print "You must specify the hostname with -n argument."
  if args.swarm:
    swarm_nodes_distribution()
  if args.container:
    remove_container(args.container)
  if args.compare:
    if not args.clustername:
      print "You must specify the cluster name with -c argument."
    else:
      compare(args.clustername)
  if args.scale:
    if not args.clustername:
      print "You must specify the cluster name with -c argument."
    elif not args.number:
      print "You must specify the scaling container number with -u argument."
    elif not args.prefix:
      print "You must specify the new container name prifix with -p argument."
    else:
      scale(args.clustername,args.number,args.prefix)

parser = argparse.ArgumentParser(description="Magpie is a Yarn-on-Docker operating tool.You can use this tool to inspect the Docker and Yarn cluster, decommisioning the nodemanagers of a host, delete the containers of a host.")
parser.add_argument('-o', '--offline', required=False, action="store_true", help='Decommisioned nodemanagers on a host.')
parser.add_argument('-d', '--delete', required=False, action="store_true", help='Delete all containers on the host no matter it is running or not.')
parser.add_argument('-s', '--scale', required=False, action="store_true", help='Scale the container number in the swarm cluster.')
parser.add_argument('-v', '--view', required=False, action="store_true", help='Show the container distribution or not, default NOT')
parser.add_argument('-c', '--cluster', required=False, dest="clustername", type=str, help='Sepecify the yarn cluster.')
parser.add_argument('-r', '--remove', required=False, dest="container", type=str, help='Sepecify the container name or ID.')
parser.add_argument('-n', '--hostname', required=False, dest="hostname", type=str, help='Sepecify the hostname.')
parser.add_argument('-p', '--prefix', required=False, dest="prefix", type=str, help='Sepecify the new container name prefix.')
parser.add_argument('-i', '--inspect', required=False,action="store_true",help='View of containers distribution on each yarn cluster.')
parser.add_argument('-w', '--swarm', required=False,action="store_true",help='View of containers distribution on swarm cluster.')
parser.add_argument('-m', '--compare', required=False,action="store_true",help='Compare the docker contianers and active nodemanagers.')
parser.add_argument('-u', '--number', required=False,dest="number", type=str, help='Specify the scaling container number.')
parser.set_defaults(func=main)
args = parser.parse_args(sys.argv[1:])
if len(sys.argv[1:])==0:
  print "Type -h for help"
args.func(args)
