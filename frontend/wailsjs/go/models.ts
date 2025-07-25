export namespace filestore {
	
	export class FileInfo {
	    id: number;
	    name: string;
	    mimeType: string;
	    timestamp: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.mimeType = source["mimeType"];
	        this.timestamp = source["timestamp"];
	        this.size = source["size"];
	    }
	}
	export class FilesInFolderResponse {
	    folderName: string;
	    files: FileInfo[];
	
	    static createFrom(source: any = {}) {
	        return new FilesInFolderResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.folderName = source["folderName"];
	        this.files = this.convertValues(source["files"], FileInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FolderInfo {
	    id: number;
	    name: string;
	    timestamp: string;
	    fileCount: number;
	
	    static createFrom(source: any = {}) {
	        return new FolderInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.timestamp = source["timestamp"];
	        this.fileCount = source["fileCount"];
	    }
	}

}

