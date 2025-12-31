export namespace main {
	
	export class WorkspaceFile {
	    name: string;
	    content: string;
	    modified: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.content = source["content"];
	        this.modified = source["modified"];
	    }
	}

}

