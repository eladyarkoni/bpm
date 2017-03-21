const express = require("express");
const app = express();

app.get('/health', function GetAppHealth(req, res){
    return res.send("I'm ok");
});

app.listen(3000, ()=>{
    console.log("Express server is up")
});

// Testing process is down by exception
// setTimeout(function(){
//     throw 'error'
// },5000);