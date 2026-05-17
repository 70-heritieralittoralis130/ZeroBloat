export namespace main {
	
	export class AppInfo {
	    packageName: string;
	    appName: string;
	    description: string;
	    isBloat: boolean;
	    isSystem: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.packageName = source["packageName"];
	        this.appName = source["appName"];
	        this.description = source["description"];
	        this.isBloat = source["isBloat"];
	        this.isSystem = source["isSystem"];
	    }
	}
	export class BloatPackage {
	    package: string;
	    app_name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new BloatPackage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.package = source["package"];
	        this.app_name = source["app_name"];
	        this.description = source["description"];
	    }
	}

}

