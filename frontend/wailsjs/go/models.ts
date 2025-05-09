export namespace filestore {
	
	export class FileInfo {
	    id: number;
	    name: string;
	    mimeType: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.mimeType = source["mimeType"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

