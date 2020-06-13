function int2array(v) {
	var a = []
	for (var i = 0; i < 32; i++) {
		if ((v&(1<<i)) != 0) {
			a.push(i+1)
		}
	}
	return a
}

function array2int(a) {
	var v
	for (var i = 0; i < a.length; i++) {
		v |= (1<<(a[i]-1))
	}
	return v
}

function str2array(v, s) {
	if (v=="") {
		return []
	}
	return v.split(s)
}

function array2str(v, s) {
	var str = ""	
	for (var i = 0; i < v.length; i++) {
		str += v[i] + s
	}
	if (v.length > 0) {
		return str.slice(0, -1)
	}
	return str
}

function get_req(url, callback) {
	var httpRequest = new XMLHttpRequest()
	httpRequest.open('GET', url, true)		
	httpRequest.send(null)
	httpRequest.onreadystatechange = function () {
		callback(httpRequest.readyState, httpRequest.status, httpRequest.responseText)			    
	};
}

function post_req(url, para, callback) {
	var httpRequest = new XMLHttpRequest()
	httpRequest.open('POST', url, true)
	httpRequest.setRequestHeader("Content-type","application/x-www-form-urlencoded")
	httpRequest.send(para)
	httpRequest.onreadystatechange = function () {
		callback(httpRequest.readyState, httpRequest.status, httpRequest.responseText)			    
	};
}

function click_tab(obj_id) {
	var ids = {		
		"t1": "p1",
		"t2": "p2",
		"t3": "p3",
		"t4": "p4",
	}
	for (id in ids) {
		document.getElementById(id).style.color = "#aaa"
		document.getElementById(ids[id]).style.display = "none"
	}
	document.getElementById(obj_id).style.color = "#0aa"
	document.getElementById(ids[obj_id]).style.display = "block"
}

function GetSelVue(id, items, val, on_change) {
	return {
		el: id,
		data: {
			options: items,    
			value: val                  
		},
		methods: {
          handle_change(value) {
          	if (on_change != null) {
          		on_change(value)
          	}            
          }
        }
	}            
}

function GetDateVue(id, sts) {
	var now = new Date().getTime()
	var ret = {
		el: id,
		data: {
			pickerOptions: {
				shortcuts: []				
			},            
			value: [new Date(now-sts[0].start), new Date(now-sts[0].end)]
		}
	}
	for (i = 0; i < sts.length; i++) {
		let start = sts[i].start
		let end = sts[i].end
		ret.data.pickerOptions.shortcuts.push({text:sts[i].text, onClick: function(picker) {
			var now = new Date().getTime()
			picker.$emit('pick', [new Date(now-start), new Date(now-end)]);
		}})
	}      
	return ret
}