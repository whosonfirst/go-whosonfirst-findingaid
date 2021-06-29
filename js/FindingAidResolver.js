/*
   
   Sample CloudFront function for resolving URIs to finding aids stored in an S3 bucket.
   This will resolve URL like:

   /123455
   /123455-alt-example.geosjon  
   /123/455/123455/123455.geojson

   to:

   /123/444/123455.json

   This function is written in JavaScript because Go is not a support CloudFront functions language yet.
*/

function handler(event) {
    
    var req = event.request;
    var uri = req.uri;
    var parts = uri.split("/");
    var last = parts.pop();
    
    var m = last.match(/^(\d+)(?:-[^\.]+)?(?:\.(?:geojson|json))?$/)
    
    if (! m){
        
        var not_found = {
            statusCode: 404,
            statusDescription: 'Not Found',
        };
        
        return not_found;
    }
    
    var id = m[1];
    var fname = id + ".json";
    
    var tmp = [];
    
    while (id.length){
	
        var part = id.substr(0, 3);
        tmp.push(part);
	id = id.substr(3);
    }
    
    var tree = tmp.join("/");
    
    var new_uri = tree + "/" + fname;
    
    var redirect = {
        statusCode: 302,
        statusDescription: 'Found',
        headers: {
            'location': { value: new_uri }
        }
    };
    
    return redirect;
}
