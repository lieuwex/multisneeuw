function makeflakeimg(width,height/*,args...*/){
	var fn=flc_makeflakedrawer.apply(null,Array.prototype.slice.call(arguments,2));
	var cvs=document.createElement("canvas");
	cvs.width=width;
	cvs.height=height;
	fn(cvs,cvs.getContext("2d"));
	var img=document.createElement("img");
	img.src=cvs.toDataURL();
	return img;
}

function flc_rotatex(x,y,ang){return x*Math.cos(ang)-y*Math.sin(ang);}
function flc_rotatey(x,y,ang){return x*Math.sin(ang)+y*Math.cos(ang);}

function flc_genbranch(depth){
	if(depth==undefined)depth=2;
	var maxsideb=depth;
	var nsideb=depth<=0?0:Math.ceil(Math.sqrt(1+8*~~(1+Math.random()*(maxsideb*(maxsideb+1)/2)))/2-0.5);
	var lines=[];
	lines.push([0,0,1,0]);
	var i,branch,offset,angle;
	for(i=0;i<nsideb;i++){
		offset=(i+1)/(nsideb+1)*3/5+Math.random()*1/10;
		angle=(Math.random()+1)*Math.PI/10;
		branch=flc_genbranch(depth-1);
		branch2=branch.map(function(line){
			return [flc_rotatex(line[0],line[1],-angle)*2/5+offset,
			        flc_rotatey(line[0],line[1],-angle)*2/5,
			        flc_rotatex(line[2],line[3],-angle)*2/5+offset,
			        flc_rotatey(line[2],line[3],-angle)*2/5]
		});
		branch=branch.map(function(line){
			return [flc_rotatex(line[0],line[1],angle)*2/5+offset,
			        flc_rotatey(line[0],line[1],angle)*2/5,
			        flc_rotatex(line[2],line[3],angle)*2/5+offset,
			        flc_rotatey(line[2],line[3],angle)*2/5]
		});
		lines=lines.concat(branch,branch2);
	}
	return lines;
}

function flc_makeflakedrawer(colour/*,args...*/){
	if(colour==undefined)colour="#aaf";
	var flake=flc_genbranch.apply(null,Array.prototype.slice.call(arguments,1));
	var nsides=7;
	return function(cvs,ctx){
		var width=cvs.width,height=cvs.height;
		var cx=width/2,cy=height/2;
		ctx.beginPath();
		ctx.strokeStyle=colour;
		var i;
		for(i=0;i<flake.length;i++){
			var x1=flake[i][0]*width/2,y1=flake[i][1]*height/2,x2=flake[i][2]*width/2,y2=flake[i][3]*height/2;
			var n,ang=2*Math.PI/nsides;
			for(n=0;n<nsides;n++){
				ctx.moveTo(flc_rotatex(x1,y1,n*ang)+cx,flc_rotatey(x1,y1,n*ang)+cy);
				ctx.lineTo(flc_rotatex(x2,y2,n*ang)+cx,flc_rotatey(x2,y2,n*ang)+cy);
			}
		}
		ctx.stroke();
	};
}