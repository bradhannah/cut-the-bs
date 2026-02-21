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
	    bullet_type: string;
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
	        this.bullet_type = source["bullet_type"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ApplicationInput {
	    company_name: string;
	    position_title: string;
	    date_applied: string;
	    fit_indicator: string;
	    resume_export_id?: number;
	    cover_letter_id?: number;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplicationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.company_name = source["company_name"];
	        this.position_title = source["position_title"];
	        this.date_applied = source["date_applied"];
	        this.fit_indicator = source["fit_indicator"];
	        this.resume_export_id = source["resume_export_id"];
	        this.cover_letter_id = source["cover_letter_id"];
	        this.notes = source["notes"];
	    }
	}
	export class BackupSettings {
	    rolling_backup_count: number;
	
	    static createFrom(source: any = {}) {
	        return new BackupSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rolling_backup_count = source["rolling_backup_count"];
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
	export class CoreExpertise {
	    id: number;
	    label: string;
	    sort_order: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new CoreExpertise(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class CoverLetter {
	    id: number;
	    title: string;
	    body_text: string;
	    file_path: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new CoverLetter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.body_text = source["body_text"];
	        this.file_path = source["file_path"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class CoverLetterInput {
	    title: string;
	    body_text: string;
	
	    static createFrom(source: any = {}) {
	        return new CoverLetterInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.body_text = source["body_text"];
	    }
	}
	export class ExportRequest {
	    template_id: string;
	    lens_id?: number;
	    summary_ids: number[];
	    master_summary_id?: number;
	    work_history_ids: number[];
	    bullet_ids: number[];
	    skill_ids: number[];
	    skill_sort_overrides: Record<number, number>;
	    academic_ids: number[];
	    certification_ids: number[];
	    descriptor_ids: number[];
	    core_expertise_ids: number[];
	
	    static createFrom(source: any = {}) {
	        return new ExportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.template_id = source["template_id"];
	        this.lens_id = source["lens_id"];
	        this.summary_ids = source["summary_ids"];
	        this.master_summary_id = source["master_summary_id"];
	        this.work_history_ids = source["work_history_ids"];
	        this.bullet_ids = source["bullet_ids"];
	        this.skill_ids = source["skill_ids"];
	        this.skill_sort_overrides = source["skill_sort_overrides"];
	        this.academic_ids = source["academic_ids"];
	        this.certification_ids = source["certification_ids"];
	        this.descriptor_ids = source["descriptor_ids"];
	        this.core_expertise_ids = source["core_expertise_ids"];
	    }
	}
	export class ImportResult {
	    records_imported: number;
	    records_skipped: number;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.records_imported = source["records_imported"];
	        this.records_skipped = source["records_skipped"];
	        this.errors = source["errors"];
	    }
	}
	export class JobApplication {
	    id: number;
	    company_name: string;
	    position_title: string;
	    date_applied: string;
	    status: string;
	    fit_indicator: string;
	    resume_export_id?: number;
	    cover_letter_id?: number;
	    notes: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new JobApplication(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.company_name = source["company_name"];
	        this.position_title = source["position_title"];
	        this.date_applied = source["date_applied"];
	        this.status = source["status"];
	        this.fit_indicator = source["fit_indicator"];
	        this.resume_export_id = source["resume_export_id"];
	        this.cover_letter_id = source["cover_letter_id"];
	        this.notes = source["notes"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class Lens {
	    id: number;
	    name: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Lens(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class LensBulletItem {
	    bullet_id: number;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new LensBulletItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bullet_id = source["bullet_id"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class LensCoreExpertiseItem {
	    core_expertise_id: number;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new LensCoreExpertiseItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.core_expertise_id = source["core_expertise_id"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class LensDescriptorItem {
	    descriptor_id: number;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new LensDescriptorItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.descriptor_id = source["descriptor_id"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class LensSkillItem {
	    skill_id: number;
	    custom_sort_order?: number;
	
	    static createFrom(source: any = {}) {
	        return new LensSkillItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skill_id = source["skill_id"];
	        this.custom_sort_order = source["custom_sort_order"];
	    }
	}
	export class LensWorkHistoryItem {
	    work_history_id: number;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new LensWorkHistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.work_history_id = source["work_history_id"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class LensSummaryItem {
	    summary_id: number;
	    sort_order: number;
	    is_master: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LensSummaryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary_id = source["summary_id"];
	        this.sort_order = source["sort_order"];
	        this.is_master = source["is_master"];
	    }
	}
	export class LensDetail {
	    id: number;
	    name: string;
	    created_at: string;
	    updated_at: string;
	    summaries: LensSummaryItem[];
	    work_history: LensWorkHistoryItem[];
	    bullets: LensBulletItem[];
	    skills: LensSkillItem[];
	    academic_ids: number[];
	    cert_ids: number[];
	    descriptors: LensDescriptorItem[];
	    core_expertise: LensCoreExpertiseItem[];
	
	    static createFrom(source: any = {}) {
	        return new LensDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	        this.summaries = this.convertValues(source["summaries"], LensSummaryItem);
	        this.work_history = this.convertValues(source["work_history"], LensWorkHistoryItem);
	        this.bullets = this.convertValues(source["bullets"], LensBulletItem);
	        this.skills = this.convertValues(source["skills"], LensSkillItem);
	        this.academic_ids = source["academic_ids"];
	        this.cert_ids = source["cert_ids"];
	        this.descriptors = this.convertValues(source["descriptors"], LensDescriptorItem);
	        this.core_expertise = this.convertValues(source["core_expertise"], LensCoreExpertiseItem);
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
	export class LensInput {
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new LensInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
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
	export class ResumeExport {
	    id: number;
	    template_id: string;
	    file_path: string;
	    summary_id?: number;
	    lens_id?: number;
	    generated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ResumeExport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.template_id = source["template_id"];
	        this.file_path = source["file_path"];
	        this.summary_id = source["summary_id"];
	        this.lens_id = source["lens_id"];
	        this.generated_at = source["generated_at"];
	    }
	}
	export class ResumeTemplate {
	    id: string;
	    name: string;
	    description: string;
	    preview_url: string;
	
	    static createFrom(source: any = {}) {
	        return new ResumeTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.preview_url = source["preview_url"];
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
	export class SkillWithTags {
	    id: number;
	    name: string;
	    category_id: number;
	    competence_level: number;
	    is_legacy: boolean;
	    created_at: string;
	    updated_at: string;
	    lens_ids: number[];
	
	    static createFrom(source: any = {}) {
	        return new SkillWithTags(source);
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
	        this.lens_ids = source["lens_ids"];
	    }
	}
	export class StatusChange {
	    id: number;
	    application_id: number;
	    from_status: string;
	    to_status: string;
	    changed_at: string;
	
	    static createFrom(source: any = {}) {
	        return new StatusChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.application_id = source["application_id"];
	        this.from_status = source["from_status"];
	        this.to_status = source["to_status"];
	        this.changed_at = source["changed_at"];
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
	    summary: string;
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
	        this.summary = source["summary"];
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
	    summary: string;
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
	        this.summary = source["summary"];
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.date_granularity_start = source["date_granularity_start"];
	        this.date_granularity_end = source["date_granularity_end"];
	    }
	}

}

