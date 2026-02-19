export namespace domain {
	
	export class AcademicCredential {
	    id: number;
	    institution: string;
	    credential_type: string;
	    field_of_study: string;
	    completion_date: string;
	    date_granularity: string;
	    sort_order: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new AcademicCredential(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.institution = source["institution"];
	        this.credential_type = source["credential_type"];
	        this.field_of_study = source["field_of_study"];
	        this.completion_date = source["completion_date"];
	        this.date_granularity = source["date_granularity"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class AcademicInput {
	    institution: string;
	    credential_type: string;
	    field_of_study: string;
	    completion_date: string;
	    date_granularity: string;
	
	    static createFrom(source: any = {}) {
	        return new AcademicInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.institution = source["institution"];
	        this.credential_type = source["credential_type"];
	        this.field_of_study = source["field_of_study"];
	        this.completion_date = source["completion_date"];
	        this.date_granularity = source["date_granularity"];
	    }
	}
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
	export class Certification {
	    id: number;
	    name: string;
	    issuing_body: string;
	    date_earned: string;
	    expiration_date: string;
	    is_active: boolean;
	    sort_order: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Certification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.issuing_body = source["issuing_body"];
	        this.date_earned = source["date_earned"];
	        this.expiration_date = source["expiration_date"];
	        this.is_active = source["is_active"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class CertificationInput {
	    name: string;
	    issuing_body: string;
	    date_earned: string;
	    expiration_date: string;
	
	    static createFrom(source: any = {}) {
	        return new CertificationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.issuing_body = source["issuing_body"];
	        this.date_earned = source["date_earned"];
	        this.expiration_date = source["expiration_date"];
	    }
	}
	export class CompetenceLevel {
	    level: number;
	    label: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new CompetenceLevel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.label = source["label"];
	        this.description = source["description"];
	    }
	}
	export class ProfessionalSummary {
	    id: number;
	    label: string;
	    body_text: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfessionalSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.body_text = source["body_text"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ProfileLink {
	    id: number;
	    label: string;
	    url: string;
	    sort_order: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.url = source["url"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ProfileLinkInput {
	    label: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileLinkInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.url = source["url"];
	    }
	}
	export class RoleDescriptor {
	    id: number;
	    title: string;
	    sort_order: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new RoleDescriptor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class Skill {
	    id: number;
	    name: string;
	    category_id: number;
	    competence_level: number;
	    is_legacy: boolean;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Skill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.category_id = source["category_id"];
	        this.competence_level = source["competence_level"];
	        this.is_legacy = source["is_legacy"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class SkillCategory {
	    id: number;
	    name: string;
	    sort_order: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class SkillCategoryWithSkills {
	    category: SkillCategory;
	    skills: Skill[];
	
	    static createFrom(source: any = {}) {
	        return new SkillCategoryWithSkills(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = this.convertValues(source["category"], SkillCategory);
	        this.skills = this.convertValues(source["skills"], Skill);
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
	export class SkillInput {
	    name: string;
	    category_id: number;
	    competence_level: number;
	    is_legacy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SkillInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.category_id = source["category_id"];
	        this.competence_level = source["competence_level"];
	        this.is_legacy = source["is_legacy"];
	    }
	}
	export class SummaryInput {
	    label: string;
	    body_text: string;
	
	    static createFrom(source: any = {}) {
	        return new SummaryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.body_text = source["body_text"];
	    }
	}
	export class UserProfile {
	    id: number;
	    full_name: string;
	    email: string;
	    phone: string;
	    location: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new UserProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.full_name = source["full_name"];
	        this.email = source["email"];
	        this.phone = source["phone"];
	        this.location = source["location"];
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

