package consts

type JobPositionStatus string

const (
	JobPositionStatusDraft     JobPositionStatus = "draft"
	JobPositionStatusPublished JobPositionStatus = "published"
)

type JobExperienceType string

const (
	JobExperienceTypeUnlimited     JobExperienceType = "unlimited"           //不限
	JobExperienceTypeFreshGraduate JobExperienceType = "fresh_graduate"      //应届生
	JobExperienceTypeUnderOneYear  JobExperienceType = "under_one_year"      //1年以下
	JobExperienceTypeOneToThree    JobExperienceType = "one_to_three_years"  //1-3年
	JobExperienceTypeThreeToFive   JobExperienceType = "three_to_five_years" //3-5年
	JobExperienceTypeFiveToTen     JobExperienceType = "five_to_ten_years"   //5-10年
	JobExperienceTypeOverTen       JobExperienceType = "over_ten_years"      //10年以上
)

// JobWorkType 工作性质类型
type JobWorkType string

const (
	JobWorkTypeFullTime    JobWorkType = "full_time"   //全职
	JobWorkTypePartTime    JobWorkType = "part_time"   //兼职
	JobWorkTypeInternship  JobWorkType = "internship"  //实习
	JobWorkTypeOutsourcing JobWorkType = "outsourcing" //外包
)

// JobEducationType 学历要求类型
type JobEducationType string

const (
	JobEducationTypeUnlimited JobEducationType = "unlimited"      //不限
	JobEducationTypeJunior    JobEducationType = "junior_college" //大专
	JobEducationTypeBachelor  JobEducationType = "bachelor"       //本科
	JobEducationTypeMaster    JobEducationType = "master"         // 硕士
	JobEducationTypeDoctor    JobEducationType = "doctor"         // 博士
)

// JobSkillType 技能类型
type JobSkillType string

const (
	JobSkillTypeRequired JobSkillType = "required" //必需技能
	JobSkillTypeBonus    JobSkillType = "bonus"    //加分技能
)

// 获取所有职位状态
func (JobPositionStatus) Values() []JobPositionStatus {
	return []JobPositionStatus{
		JobPositionStatusDraft,
		JobPositionStatusPublished,
	}
}

// 验证是否为有效值
func (j JobPositionStatus) IsValid() bool {
	for _, v := range j.Values() {
		if j == v {
			return true
		}
	}
	return false
}

// 获取所有工作经验类型
func (JobExperienceType) Values() []JobExperienceType {
	return []JobExperienceType{
		JobExperienceTypeUnlimited,
		JobExperienceTypeFreshGraduate,
		JobExperienceTypeUnderOneYear,
		JobExperienceTypeOneToThree,
		JobExperienceTypeThreeToFive,
		JobExperienceTypeFiveToTen,
		JobExperienceTypeOverTen,
	}
}

// 验证是否为有效值
func (j JobExperienceType) IsValid() bool {
	for _, v := range j.Values() {
		if j == v {
			return true
		}
	}
	return false
}

// 获取所有工作类型
func (JobWorkType) Values() []JobWorkType {
	return []JobWorkType{
		JobWorkTypeFullTime,
		JobWorkTypePartTime,
		JobWorkTypeInternship,
		JobWorkTypeOutsourcing,
	}
}

// 验证是否为有效值
func (j JobWorkType) IsValid() bool {
	for _, v := range j.Values() {
		if j == v {
			return true
		}
	}
	return false
}

// 获取所有学历类型
func (JobEducationType) Values() []JobEducationType {
	return []JobEducationType{
		JobEducationTypeUnlimited,
		JobEducationTypeJunior,
		JobEducationTypeBachelor,
		JobEducationTypeMaster,
		JobEducationTypeDoctor,
	}
}

// 验证是否为有效值
func (j JobEducationType) IsValid() bool {
	for _, v := range j.Values() {
		if j == v {
			return true
		}
	}
	return false
}

// 获取所有技能类型
func (JobSkillType) Values() []JobSkillType {
	return []JobSkillType{
		JobSkillTypeRequired,
		JobSkillTypeBonus,
	}
}

// 验证是否为有效值
func (j JobSkillType) IsValid() bool {
	for _, v := range j.Values() {
		if j == v {
			return true
		}
	}
	return false
}
