var DELIMITER="⛄️";
var ID_LENGTH = 10;

var errorMessages = {
	'connection-closed': 'Connection unexpectedly closed.',
	'invalid-side': null,
};
var flakeColours={
	"self": "#aaf",
	"L": "#5f5",
	"R": "#ffabab"
};

var ws,wssendqueue=[];
var noUpdateFlakes=false;
var errored=false;

var FLAKEWID=40,FLAKEHEI=40;
var flakes=[];

var globalCanvas,globalContext;

var comboScore=0;

var uniqid = (function () {
	var chars = [];
	for (var i = 33; i <= 126; i++) chars.push(String.fromCharCode(i));

	return function () {
		var arr = new Array(ID_LENGTH);
		for (var i = 0; i < ID_LENGTH; i++) {
			arr[i] = chars[~~(Math.random() * (chars.length + 1))]
		}
		return arr.join('');
	};
}());

function Flake(){
	this.id=uniqid();
	this.img=arguments[0]; //the flake canvas
	this.pos=arguments[1]; //position
	this.speed=arguments[2]; //speed of the flake
	this.rot=0; //rotation angle
	this.rotspeed=arguments[3]; //rotation speed
	this.dir=arguments[4]; //current movement direction
	//this.opacity=arguments[5]; //flake opacity
	this.dragged=false; //whether dragged by mouse; inhibits motion
	this.combo=arguments[5]?arguments[5]:0; //amount of times thrown across screen boundary
	this.alreadyDragged=false;
}

Flake.prototype.toJSON = function () {
	return {
		id: this.id,
		pos: this.pos,
		speed: this.speed,
		rot: this.rot,
		rotspeed: this.rotspeed,
		dir: this.dir,
		combo: this.combo,
		//opacity: this.opacity,
	};
};

function addflake(bodycr){
	addflakeparams({
		img: makeflakeimg(FLAKEWID,FLAKEHEI,flakeColours.self),
		pos: [ Math.random()*bodycr.width, -FLAKEHEI ],
		speed: 120,
		rotspeed: (Math.random()-0.5)*Math.PI/100,
		dir: (Math.random()*2-1)*Math.PI/6,
		//opacity: Math.random()*0.3+0.7,
	});
}

function drawflake(flake){
	globalContext.save();
	globalContext.translate(flake.pos[0]+FLAKEWID*2,flake.pos[1]+FLAKEHEI*2);
	globalContext.rotate(flake.rot);
	globalContext.translate(-FLAKEWID/2,-FLAKEHEI/2);
	globalContext.drawImage(flake.img,0,0);
	globalContext.restore();
	if(flake.combo>=2){
		globalContext.font="bold 13px sans-serif";
		globalContext.textAlign="center";
		globalContext.textBaseline="middle";
		globalContext.fillStyle="#55f";
		globalContext.fillText(flake.combo.toString(),flake.pos[0]+FLAKEWID*2,flake.pos[1]+FLAKEHEI*2);
	}
}

function addflakeparams(options){
	var img=options.img;
	var flake=new Flake(
		img,
		options.pos,
		options.speed,
		options.rotspeed,
		options.dir,
		options.combo
		//options.opacity
	);
	flakes.push(flake);
	drawflake(flake);
}

function modulo(a,b){
	return (a%b+b)%b;
}

var updateFlakesPrevTimestamp,cumulativeDiff=0,fps=60;
var timestamphist=[];
function updateFlakes(timestamp){
	var bodycr=document.body.getBoundingClientRect();
	var difftime;
	var i;
	if(timestamphist.length==60)timestamphist.shift();
	timestamphist.push(timestamp/1000);
	fps=60/(timestamphist[timestamphist.length-1]-timestamphist[0]);
	if(updateFlakesPrevTimestamp!=null){
		difftime=(timestamp-updateFlakesPrevTimestamp)/1000;
		cumulativeDiff+=difftime;
		if(cumulativeDiff>=1/10){
			addflake(bodycr);
			cumulativeDiff=0;
		}
	} else difftime=1/60;
	updateFlakesPrevTimestamp=timestamp;

	globalContext.clearRect(0,0,globalCanvas.width,globalCanvas.height);

	var shouldremove;
	for(i=0;i<flakes.length;i++){
		drawflake(flakes[i]);
		if(flakes[i].dragged)continue;
		flakes[i].pos[0]+=flakes[i].speed*difftime*Math.sin(flakes[i].dir);
		flakes[i].pos[1]+=flakes[i].speed*difftime*Math.cos(flakes[i].dir);
		shouldremove=false;
		if(flakes[i].pos[1]>=bodycr.height+FLAKEHEI/2){
			shouldremove=true;
		} else if(flakes[i].pos[0]<=-FLAKEWID/2){
			if(modulo(flakes[i].dir,2*Math.PI)>Math.PI){
				flakes[i].combo++;
				sendMsg(["L",flakes[i].pos[1]/bodycr.height,JSON.stringify(flakes[i])].join(DELIMITER));
			}
			shouldremove=true;
		} else if(flakes[i].pos[0]>=bodycr.width+FLAKEWID/2){
			if(modulo(flakes[i].dir,2*Math.PI)<Math.PI){
				flakes[i].combo++;
				sendMsg(["R",flakes[i].pos[1]/bodycr.height,JSON.stringify(flakes[i])].join(DELIMITER));
			}
			shouldremove=true;
		}
		if(shouldremove){
			delete flakes[i].img;
			flakes.splice(i,1);
			i--;
			continue;
		}

		flakes[i].rot+=flakes[i].rotspeed;
	}
	if(noUpdateFlakes)return;
	window.requestAnimationFrame(updateFlakes);
}


function notifyClosed(){
	var bodycr=document.body.getBoundingClientRect();
	var minsz=Math.min(bodycr.width,bodycr.height);
	var img=makeflakeimg(minsz,minsz,flakeColours.self,3);
	img.setAttribute("style","position:fixed;left:50%;top:50%;transform:translateY(-50%) translateX(-50%);");
	document.body.appendChild(img);

	noUpdateFlakes=true;
	document.body.classList.add("black");
}

function notifyError(errCode){
	if(errored)return;
	errored=true;

	var message = errorMessages[errCode];
	if(message==null)message = errCode;

	var msgdiv = document.createElement('div');
	msgdiv.className = 'error-message';
	msgdiv.appendChild(document.createTextNode(message));
	document.body.appendChild(msgdiv);
}

function updateComboCounter(){
	document.getElementById("comboCounter").innerHTML=comboScore;
}


function connect(){
	ws=new WebSocket("ws://"+location.host+"/ws");
	var noreplypings=0;
	var pinginterval;
	ws.addEventListener("open",function(){
		wssendqueue.forEach(function(msg){
			sendMsg(msg);
		});
		wssendqueue=[];
		pinginterval=setInterval(function(){
			if(ws.readyState!=ws.OPEN||noreplypings>=1){
				ws.close();
				clearInterval(pinginterval);
				notifyClosed();
				notifyError("connection-closed");
				return;
			}
			sendMsg("ping"+DELIMITER+"pong");
			noreplypings++;
		},10000);
	});
	ws.addEventListener("message",function(msg){
		noreplypings=0;
		msg=msg.data.slice(0, -1);
		//console.log(msg);
		if(msg.slice(0,4)=="yolo"){
			ws.close();
			notifyClosed();

			notifyError(msg.split(DELIMITER)[1]);
			return;
		} else if(msg.slice(0,4)=="ping"){
			sendMsg("pong"+DELIMITER+"ping");
		} else if(msg.slice(0,4)=="pong"){
			lastPong=new Date().getTime();
		} else if(msg.slice(0,5)=="index"){
			var counter = document.getElementById('indexCounter');
			var splitted = msg.split(DELIMITER);
			document.body.classList[splitted[2]>1?'remove':'add']('waiting');
			counter.innerText = splitted[1] + '/' + splitted[2];
		}
		msg=msg.split(DELIMITER);
		if(msg[0]=="B")msg.shift(); //the B of broadcast is uninteresting
		if(msg[0]=="score"){
			comboScore=+msg[1];
			updateComboCounter();
		}
		if(msg[0]!="L"&&msg[0]!="R")return;
		var side=msg[0];
		var y=parseFloat(msg[1]);
		msg=JSON.parse(msg[2]);
		var bodywidth=document.body.getBoundingClientRect().width;
		// console.log(side,y,msg);
		addflakeparams({
			img: makeflakeimg(FLAKEWID,FLAKEHEI,flakeColours[side]),
			pos: [side=="L"?bodywidth-1:-FLAKEWID/2+1, y*document.body.getBoundingClientRect().height],
			speed: msg.speed,
			rotspeed: msg.rotspeed,
			dir: msg.dir,
			combo: msg.combo,
			//opacity: msg.opacity,
		});
		if(msg.combo>=2){
			comboScore+=msg.combo;
			updateComboCounter();
			sendMsg("addscore" + DELIMITER + msg.combo);
		}
		// console.log(flakes[flakes.length-1]);
	});
	ws.addEventListener("close",function(){
		clearInterval(pinginterval);
		notifyClosed();
		notifyError('connection-closed');
	});
}

function sendMsg(msg){
	if(!ws||ws.readyState!=ws.OPEN)wssendqueue.push(msg);
	else {
		ws.send(msg+"\n");
	}
}


function attachMouseListeners(){
	var selected=null,offset;
	var cvscr;
	var draghist;
	globalCanvas.addEventListener("mousedown",function(ev){
		cvscr=globalCanvas.getBoundingClientRect();
		if(selected!=null){
			selected.dragged=false;
		}
		selected=offset=null;
		var x=ev.clientX,y=ev.clientY;
		var i;
		for(i=0;i<flakes.length;i++){
			if((flakes[i].pos[0]-x)*(flakes[i].pos[0]-x)+(flakes[i].pos[1]-y)*(flakes[i].pos[1]-y)<=FLAKEWID*FLAKEWID/4){
				break;
			}
		}
		if(i==flakes.length)return;
		flakes[i].dragged=true;
		selected=flakes[i];
		if(selected.alreadyDragged)selected.combo=0;
		selected.alreadyDragged=true;
		offset=[x-flakes[i].pos[0],y-flakes[i].pos[1]];
		draghist=[];
	});
	globalCanvas.addEventListener("mousemove",function(ev){
		if(selected==null)return;
		var x=ev.clientX,y=ev.clientY;
		while(draghist.length>=7)draghist.shift();
		draghist.push([x,y,new Date().getTime()/1000]);
		x-=offset[0];
		y-=offset[1];
		selected.pos[0]=x;
		selected.pos[1]=y;
	});
	globalCanvas.addEventListener("mouseup",function(ev){
		if(selected==null)return;
		selected.dragged=false;
		if(draghist.length<2){
			selected.speed/=5; //short click is a slowdown
			selected=offset=null;
			return;
		}
		var dir,dirx=0,diry=0;
		var speed=0;
		var timedelta=0;
		var nsteps=Math.min(draghist.length-1,6);
		for(var i=0;i<nsteps;i++){
			dir=Math.atan2(draghist[i+1][1]-draghist[i][1],draghist[i+1][0]-draghist[i][0]);
			dirx+=Math.cos(dir); diry+=Math.sin(dir);
			speed+=Math.sqrt((draghist[i][0]-draghist[i+1][0])*(draghist[i][0]-draghist[i+1][0])+(draghist[i][1]-draghist[i+1][1])*(draghist[i][1]-draghist[i+1][1]));
			timedelta+=draghist[i+1][2]-draghist[i][2];
		}
		dir=Math.atan2(diry,dirx);
		// console.log("timedelta =",timedelta);
		// console.log(speed, speed/nsteps, speed/nsteps/timedelta, speed/nsteps/timedelta/fps);
		speed=speed/nsteps/timedelta;
		// console.log(dir,speed);
		selected.dir=Math.PI/2-dir;
		selected.speed=speed;
		selected=offset=null;
	});
}


function updateCanvasSize(){
	var bodycr=document.body.getBoundingClientRect();
	globalCanvas.width=bodycr.width+4*FLAKEWID;
	globalCanvas.height=bodycr.height+4*FLAKEHEI;
}

window.addEventListener("load",function(){
	globalCanvas=document.getElementById("cvs");
	globalContext=globalCanvas.getContext("2d");
	updateCanvasSize();
	updateComboCounter();
	connect();
	sendMsg(location.pathname.slice(1));

	attachMouseListeners();

	window.requestAnimationFrame(updateFlakes);
});

window.addEventListener("resize",function(){
	updateCanvasSize();
});
