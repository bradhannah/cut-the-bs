export namespace domain {
	
	export class AchievementBullet {
	    id: number;
	    work_history_id: number;
	    text: string;
	    sort_order: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new AchievementBullet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.work_history_id = source["work_history_id"];
	        this.text = source["text"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class WorkHistoryEntry {
	    id: number;
	    employer_name: string;
	    job_title: string;
	    start_date: string;
	    end_date: string;
	    date_granularity_start: string;
	    date_granularity_end: string;
	    sort_order: number;
	    bullets: AchievementBullet[];
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkHistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.employer_name = source["employer_name"];
	        this.job_title = source["job_title"];
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.date_granularity_start = source["date_granularity_start"];
	        this.date_granularity_end = source["date_granularity_end"];
	        this.sort_order = source["sort_order"];
	        this.bullets = this.convertValues(source["bullets"], AchievementBullet);
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
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
	export class WorkHistoryInput {
	    employer_name: string;
	    job_title: string;
	    start_date: string;
	    end_date: string;
	    date_granularity_start: string;
	    date_granularity_end: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkHistoryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.employer_name = source["employer_name"];
	        this.job_title = source["job_title"];
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.date_granularity_start = source["date_granularity_start"];
	        this.date_granularity_end = source["date_granularity_end"];
	    }
	}

}

