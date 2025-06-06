package models

// 录取数据表结构
type AdmissionData struct {
	ID                       int64   `json:"id" ch:"id"`                                           // 自增ID
	Year                     int     `json:"year" ch:"year"`                                       // 年份
	Province                 string  `json:"province" ch:"province"`                               // 省份
	CollegeName              string  `json:"college_name" ch:"college_name"`                       // 院校名称
	CollegeCode              string  `json:"college_code" ch:"college_code"`                       // 院校代码
	SpecialInterestGroupCode string  `json:"special_interest_group_code" ch:"special_interest_group_code"` // 专业组代码
	ProfessionalName         string  `json:"professional_name" ch:"professional_name"`             // 专业名称
	ClassDemand              string  `json:"class_demand" ch:"class_demand"`                       // 选科要求
	LowestPoints             int64   `json:"lowest_points" ch:"lowest_points"`                     // 录取最低分
	LowestRank               int64   `json:"lowest_rank" ch:"lowest_rank"`                         // 录取最低位次
	Description              string  `json:"description" ch:"description"`                         // 备注
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

type List struct {
	ClassDemand              *string `json:"class_demand,omitempty"`
	ColledgeCode             *string `json:"colledge_code,omitempty"`
	ColledgeName             *string `json:"colledge_name,omitempty"`
	Description              *string `json:"description,omitempty"`
	ID                       *int64  `json:"id,omitempty"`
	LowestPoints             *int64  `json:"lowest_points,omitempty"`
	LowestRank               *int64  `json:"lowest_rank,omitempty"`
	ProfessionalName         string  `json:"professional_name"`
	SpecialInterestGroupCode *string `json:"special_interest_group_code,omitempty"`
}

// 位次查询结果
type RankResponse struct {
	Code int64 `json:"code"`
	Msg  string `json:"msg"`
	Rank int64 `json:"rank"`
	Year int   `json:"year"`
} 