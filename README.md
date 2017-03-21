# Bulk Process Manager
**BPM is a Process manager for [nodejs](http://nodejs.org) projects**

**Author: [Elad Yarkoni](http://eladyarkoni.com)**

<div style="text-align: center;">
    <img src="https://nodejs.org/static/images/logos/nodejs-new-pantone-black.png" height="100">
    <img src="https://gquintana.github.io//images/logos/golang.png" height="100">
    <img src="https://roma-kvs.org/images/benchmark/leveldb-logo.png" height="100">
</div>
<br/>

BPM is a production process manager for Node.js applications developed in [go language](http://golang.org).
It allows you to keep applications alive forever and to run them in a cluster mode.

BPM gives you a set of tools to manage your node production processes through command line or even remotly.

## Behind The Scenes
BPM is using a local http server as a god of all nodejs processes.  
The bpm command will start the server if its not started yet and communicate with it through rest api calls.  
The BPM local http server is using Google LevelDB, a fast key-value storage library, to store your projects data and status.  
Eventually, The bpm http server monitors your running nodejs processes. you can use the http server remotly to take a full control over your nodejs projects.  

Any bpm command line will do the following:
1. Starts the local http server is its not started yet
2. Makes a rest api call
3. Gets a Response
4. Prints the response to command line

## Install
```
$ go get -u github.com/eladyarkoni/bpm
```

## Usage
### Adding a new nodejs project to BPM
This command adds the nodejs project working directory to BPM.  
```
$ bpm add <node_project_directory>
```

The node project must have the package.json file with the main script configured.  
BPM is using the main script to start the process.  

```
{
    "name": "node-project-name",
    "description": "node project description...",
    ...
    ...
    "main": "index.js"
}
```

### Start Node Project
This command starts the nodejs project processes. 
```
$ bpm start <package_name> [cluster_processes_number]
```

**cluster_processes_number**: The number of processes to start the project in cluster mode.  
  
* If cluster_processes_number is not defined or 0, then, the node project will be started in normal mode.  

### Stop Node Project
This command stops the nodejs project processes. 
```
$ bpm stop <package_name>
```

### Get Status
This command gets the status of all nodejs projects that are managed in BPM.
```
$ bpm status
```

## Development Roadmap
BPM is going to be the ultimate solution for managing NodeJS projects on production environment.  
Here are some of the features that are going to be developed in the near future:
1. Process resources monitoring (Memory, CPU...)
2. Autoscaling support
3. Alerts configuration for DevOps
4. Deploy: Deploy nodejs projects to remote servers.
5. BPM Center Orchestrator: Manage nodejs projects over multiple servers
 
And more...

## License
BSD-2-Clause.


