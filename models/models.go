package models

// 录取数据表结构 - 新的ClickHouse表结构
type AdmissionHubeiWide struct {
	ID                    uint32 `json:"id" ch:"id"`
	SchoolCode            string `json:"school_code" ch:"school_code"`
	SchoolName            string `json:"school_name" ch:"school_name"`
	MajorCode             string `json:"major_code" ch:"major_code"`
	MajorName             string `json:"major_name" ch:"major_name"`
	MajorGroupCode        string `json:"major_group_code" ch:"major_group_code"`
	SourceProvince        string `json:"source_province" ch:"source_province"`
	SchoolProvince        string `json:"school_province" ch:"school_province"`
	SchoolCity            string `json:"school_city" ch:"school_city"`
	AdmissionBatch        string `json:"admission_batch" ch:"admission_batch"`
	SubjectCategory       string `json:"subject_category" ch:"subject_category"`
	RequirePhysics        bool   `json:"require_physics" ch:"require_physics"`
	RequireChemistry      bool   `json:"require_chemistry" ch:"require_chemistry"`
	RequireBiology        bool   `json:"require_biology" ch:"require_biology"`
	RequirePolitics       bool   `json:"require_politics" ch:"require_politics"`
	RequireHistory        bool   `json:"require_history" ch:"require_history"`
	RequireGeography      bool   `json:"require_geography" ch:"require_geography"`
	SubjectRequirementRaw string `json:"subject_requirement_raw" ch:"subject_requirement_raw"`
	SchoolType            string `json:"school_type" ch:"school_type"`
	SchoolOwnership       string `json:"school_ownership" ch:"school_ownership"`
	SchoolAuthority       string `json:"school_authority" ch:"school_authority"`
	SchoolLevel           string `json:"school_level" ch:"school_level"`
	SchoolTags            string `json:"school_tags" ch:"school_tags"`
	EducationLevel        string `json:"education_level" ch:"education_level"`
	MajorDescription      string `json:"major_description" ch:"major_description"`
	StudyYears            uint8  `json:"study_years" ch:"study_years"`
	TuitionFee            uint32 `json:"tuition_fee" ch:"tuition_fee"`
	IsNewMajor            bool   `json:"is_new_major" ch:"is_new_major"`
	MinScore2024          uint16 `json:"min_score_2024" ch:"min_score_2024"`
	MinRank2024           uint32 `json:"min_rank_2024" ch:"min_rank_2024"`
	EnrollmentPlan2024    uint16 `json:"enrollment_plan_2024" ch:"enrollment_plan_2024"`
	IsScience             bool   `json:"is_science" ch:"is_science"`
	IsEngineering         bool   `json:"is_engineering" ch:"is_engineering"`
	IsMedical             bool   `json:"is_medical" ch:"is_medical"`
	IsEconomicsMgmtLaw    bool   `json:"is_economics_mgmt_law" ch:"is_economics_mgmt_law"`
	IsLiberalArts         bool   `json:"is_liberal_arts" ch:"is_liberal_arts"`
	IsDesignArts          bool   `json:"is_design_arts" ch:"is_design_arts"`
	IsLanguage            bool   `json:"is_language" ch:"is_language"`
}

// 旧的录取数据表结构（保持兼容性）
type AdmissionData struct {
	ID                       int64  `json:"id" ch:"id"`                                                   // 自增ID
	Year                     int    `json:"year" ch:"year"`                                               // 年份
	Province                 string `json:"province" ch:"province"`                                       // 省份
	Batch                    string `json:"batch" ch:"batch"`                                             // 批次
	SubjectType              string `json:"subject_type" ch:"subject_type"`                               // 科类
	ClassDemand              string `json:"class_demand" ch:"class_demand"`                               // 选科要求
	CollegeCode              string `json:"college_code" ch:"college_code"`                               // 院校代码
	SpecialInterestGroupCode string `json:"special_interest_group_code" ch:"special_interest_group_code"` // 专业组代码
	CollegeName              string `json:"college_name" ch:"college_name"`                               // 院校名称
	ProfessionalCode         string `json:"professional_code" ch:"professional_code"`                     // 专业代码
	ProfessionalName         string `json:"professional_name" ch:"professional_name"`                     // 专业名称
	LowestPoints             int64  `json:"lowest_points" ch:"lowest_points"`                             // 录取最低分
	LowestRank               int64  `json:"lowest_rank" ch:"lowest_rank"`                                 // 录取最低位次
	Description              string `json:"description" ch:"description"`                                 // 备注
}

// API响应结构
type Response struct {
	Code int64  `json:"code"`
	Data Data   `json:"data,omitempty"`
	Msg  string `json:"msg"`
	Rank int64  `json:"rank,omitempty"`
	Year int    `json:"year,omitempty"`
}

type Data struct {
	Conf *Conf  `json:"conf,omitempty"`
	List []List `json:"list"`
}

type Conf struct {
	Page        int64 `json:"page"`
	PageSize    int64 `json:"page_size"`
	TotalNumber int64 `json:"total_number"`
	TotalPage   int64 `json:"total_page"`
}

// 新的报表响应结构
type List struct {
	ID                       *uint64 `json:"id,omitempty"`
	ColledgeName             *string `json:"colledge_name,omitempty"`
	ColledgeCode             *string `json:"colledge_code,omitempty"`
	SpecialInterestGroupCode *string `json:"special_interest_group_code,omitempty"`
	ClassDemand              *string `json:"class_demand,omitempty"`
	ColledgeProvince         *string `json:"colledge_province,omitempty"`
	ColledgeCity             *string `json:"colledge_city,omitempty"`
	ColledgeOwnership        *string `json:"colledge_ownership,omitempty"`
	ColledgeType             *string `json:"colledge_type,omitempty"`
	ColledgeAuthority        *string `json:"colledge_authority,omitempty"`
	ColledgeLevel            *string `json:"colledge_level,omitempty"`
	ColledgeTags             *string `json:"colledge_tags,omitempty"`
	EducationLevel           *string `json:"education_level,omitempty"`
	MajorDescription         *string `json:"major_description,omitempty"`
	TuitionFee               *uint32 `json:"tuition_fee,omitempty"`
	IsNewMajor               *bool   `json:"is_new_major,omitempty"`
	// 保持兼容性的字段
	Description      *string `json:"description,omitempty"`
	LowestPoints     *int64  `json:"lowest_points,omitempty"`
	LowestRank       *int64  `json:"lowest_rank,omitempty"`
	ProfessionalName string  `json:"professional_name"`
}

// 位次查询结果
type RankResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Rank int64  `json:"rank"`
	Year int    `json:"year"`
}
