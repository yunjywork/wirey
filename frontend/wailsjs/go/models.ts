export namespace config {
	
	export class ConnectionSettings {
	    connectTimeout: number;
	    readTimeout: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectTimeout = source["connectTimeout"];
	        this.readTimeout = source["readTimeout"];
	    }
	}
	export class AppSettings {
	    defaultCharset: string;
	    connectionSettings?: ConnectionSettings;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultCharset = source["defaultCharset"];
	        this.connectionSettings = this.convertValues(source["connectionSettings"], ConnectionSettings);
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
	export class CaseOrder {
	    id: string;
	    order: number;
	
	    static createFrom(source: any = {}) {
	        return new CaseOrder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.order = source["order"];
	    }
	}
	export class ScriptConfig {
	    setupScript: string;
	    setupEnabled: boolean;
	    preSendScript: string;
	    preSendEnabled: boolean;
	    postRecvScript: string;
	    postRecvEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScriptConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.setupScript = source["setupScript"];
	        this.setupEnabled = source["setupEnabled"];
	        this.preSendScript = source["preSendScript"];
	        this.preSendEnabled = source["preSendEnabled"];
	        this.postRecvScript = source["postRecvScript"];
	        this.postRecvEnabled = source["postRecvEnabled"];
	    }
	}
	export class FramingConfig {
	    mode: string;
	    delimiter?: string;
	    lengthEncoding?: string;
	    lengthOffset?: number;
	    lengthBytes?: number;
	    endian?: string;
	    includeHeader?: boolean;
	    lengthMode?: string;
	    fixedSize?: number;
	
	    static createFrom(source: any = {}) {
	        return new FramingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.delimiter = source["delimiter"];
	        this.lengthEncoding = source["lengthEncoding"];
	        this.lengthOffset = source["lengthOffset"];
	        this.lengthBytes = source["lengthBytes"];
	        this.endian = source["endian"];
	        this.includeHeader = source["includeHeader"];
	        this.lengthMode = source["lengthMode"];
	        this.fixedSize = source["fixedSize"];
	    }
	}
	export class Collection {
	    name: string;
	    description?: string;
	    createdAt: string;
	    updatedAt: string;
	    order: number;
	    sharedFraming?: FramingConfig;
	    sharedCharset?: string;
	    sharedConnectionSettings?: ConnectionSettings;
	    notes?: string;
	    sharedScriptConfig?: ScriptConfig;
	    sharedVariables?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Collection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.order = source["order"];
	        this.sharedFraming = this.convertValues(source["sharedFraming"], FramingConfig);
	        this.sharedCharset = source["sharedCharset"];
	        this.sharedConnectionSettings = this.convertValues(source["sharedConnectionSettings"], ConnectionSettings);
	        this.notes = source["notes"];
	        this.sharedScriptConfig = this.convertValues(source["sharedScriptConfig"], ScriptConfig);
	        this.sharedVariables = source["sharedVariables"];
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
	export class CollectionOrder {
	    name: string;
	    order: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionOrder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.order = source["order"];
	    }
	}
	export class SavedCase {
	    id: string;
	    name: string;
	    protocol: string;
	    host: string;
	    port: number;
	    createdAt: string;
	    updatedAt: string;
	    order: number;
	    framing?: FramingConfig;
	    charset?: string;
	    connectionSettings?: ConnectionSettings;
	    draftMessage?: string;
	    draftFormat?: string;
	    useVariables: boolean;
	    notes?: string;
	    scriptConfig?: ScriptConfig;
	    localVariables?: Record<string, any>;
	    postRecvSample?: string;
	
	    static createFrom(source: any = {}) {
	        return new SavedCase(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.protocol = source["protocol"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.order = source["order"];
	        this.framing = this.convertValues(source["framing"], FramingConfig);
	        this.charset = source["charset"];
	        this.connectionSettings = this.convertValues(source["connectionSettings"], ConnectionSettings);
	        this.draftMessage = source["draftMessage"];
	        this.draftFormat = source["draftFormat"];
	        this.useVariables = source["useVariables"];
	        this.notes = source["notes"];
	        this.scriptConfig = this.convertValues(source["scriptConfig"], ScriptConfig);
	        this.localVariables = source["localVariables"];
	        this.postRecvSample = source["postRecvSample"];
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
	export class CollectionWithCases {
	    collection: Collection;
	    cases: SavedCase[];
	
	    static createFrom(source: any = {}) {
	        return new CollectionWithCases(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.collection = this.convertValues(source["collection"], Collection);
	        this.cases = this.convertValues(source["cases"], SavedCase);
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
	
	
	

}

export namespace echo {
	
	export class LogEntry {
	    id: string;
	    timestamp: number;
	    direction: string;
	    remoteAddr: string;
	    data: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.direction = source["direction"];
	        this.remoteAddr = source["remoteAddr"];
	        this.data = source["data"];
	        this.size = source["size"];
	    }
	}
	export class Status {
	    running: boolean;
	    port: number;
	    protocol: string;
	    address: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.address = source["address"];
	    }
	}

}

export namespace main {
	
	export class PostRecvDryRunRequest {
	    message: string;
	    caseId: string;
	    collectionName: string;
	    caseScriptConfig?: config.ScriptConfig;
	    collectionVariables: Record<string, any>;
	    caseVariables: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new PostRecvDryRunRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.caseId = source["caseId"];
	        this.collectionName = source["collectionName"];
	        this.caseScriptConfig = this.convertValues(source["caseScriptConfig"], config.ScriptConfig);
	        this.collectionVariables = source["collectionVariables"];
	        this.caseVariables = source["caseVariables"];
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
	export class ScriptDryRunRequest {
	    message: string;
	    caseId: string;
	    collectionName: string;
	    caseScriptConfig?: config.ScriptConfig;
	    collectionVariables: Record<string, any>;
	    caseVariables: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ScriptDryRunRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.caseId = source["caseId"];
	        this.collectionName = source["collectionName"];
	        this.caseScriptConfig = this.convertValues(source["caseScriptConfig"], config.ScriptConfig);
	        this.collectionVariables = source["collectionVariables"];
	        this.caseVariables = source["caseVariables"];
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
	export class ScriptDryRunResponse {
	    result: string;
	    logs: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ScriptDryRunResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.result = source["result"];
	        this.logs = source["logs"];
	        this.error = source["error"];
	    }
	}
	export class ScriptProcessRequest {
	    message: string;
	    caseId: string;
	    collectionName: string;
	    collectionScriptConfig?: config.ScriptConfig;
	    caseScriptConfig?: config.ScriptConfig;
	    collectionVariables: Record<string, any>;
	    caseVariables: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ScriptProcessRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.caseId = source["caseId"];
	        this.collectionName = source["collectionName"];
	        this.collectionScriptConfig = this.convertValues(source["collectionScriptConfig"], config.ScriptConfig);
	        this.caseScriptConfig = this.convertValues(source["caseScriptConfig"], config.ScriptConfig);
	        this.collectionVariables = source["collectionVariables"];
	        this.caseVariables = source["caseVariables"];
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

}

