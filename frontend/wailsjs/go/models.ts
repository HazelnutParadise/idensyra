export namespace igonb {
	
	export class CellResult {
	    index: number;
	    language: string;
	    output: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new CellResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.language = source["language"];
	        this.output = source["output"];
	        this.error = source["error"];
	    }
	}

}

export namespace main {
	
	export class WorkspaceFile {
	    name: string;
	    content: string;
	    modified: boolean;
	    size: number;
	    tooLarge: boolean;
	    isBinary: boolean;
	    isDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.content = source["content"];
	        this.modified = source["modified"];
	        this.size = source["size"];
	        this.tooLarge = source["tooLarge"];
	        this.isBinary = source["isBinary"];
	        this.isDir = source["isDir"];
	    }
	}

}

