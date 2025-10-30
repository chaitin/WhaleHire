package domain

import "github.com/chaitin/WhaleHire/backend/consts"

// convertToBasicMatchDetail 转换基本信息匹配详情
func convertToBasicMatchDetail(data map[string]interface{}) *BasicMatchDetail {
	if data == nil {
		return nil
	}

	detail := &BasicMatchDetail{}

	if score, ok := data["score"].(float64); ok {
		detail.Score = score
	}

	if subScores, ok := data["sub_scores"].(map[string]interface{}); ok {
		detail.SubScores = make(map[string]float64)
		for k, v := range subScores {
			if f, ok := v.(float64); ok {
				detail.SubScores[k] = f
			}
		}
	}

	if evidence, ok := data["evidence"].([]interface{}); ok {
		detail.Evidence = make([]string, 0, len(evidence))
		for _, e := range evidence {
			if s, ok := e.(string); ok {
				detail.Evidence = append(detail.Evidence, s)
			}
		}
	}

	if notes, ok := data["notes"].(string); ok {
		detail.Notes = notes
	}

	return detail
}

// 辅助转换函数

// convertToMatchedSkill 转换匹配的技能
func convertToMatchedSkill(data map[string]interface{}) *MatchedSkill {
	if data == nil {
		return nil
	}

	skill := &MatchedSkill{}

	if jobSkillID, ok := data["job_skill_id"].(string); ok {
		skill.JobSkillID = jobSkillID
	}

	if resumeSkillID, ok := data["resume_skill_id"].(string); ok {
		skill.ResumeSkillID = resumeSkillID
	}

	if matchType, ok := data["match_type"].(string); ok {
		skill.MatchType = matchType
	}

	if llmScore, ok := data["llm_score"].(float64); ok {
		skill.LLMScore = llmScore
	}

	if proficiencyGap, ok := data["proficiency_gap"].(float64); ok {
		skill.ProficiencyGap = proficiencyGap
	}

	if score, ok := data["score"].(float64); ok {
		skill.Score = score
	}

	if llmAnalysis, ok := data["llm_analysis"].(map[string]interface{}); ok {
		skill.LLMAnalysis = convertToSkillItemAnalysis(llmAnalysis)
	}

	return skill
}

// convertToJobSkill 转换职位技能
func convertToJobSkill(data map[string]interface{}) *JobSkill {
	if data == nil {
		return nil
	}

	skill := &JobSkill{}

	if id, ok := data["id"].(string); ok {
		skill.ID = id
	}

	if jobID, ok := data["job_id"].(string); ok {
		skill.JobID = jobID
	}

	if skillID, ok := data["skill_id"].(string); ok {
		skill.SkillID = skillID
	}

	if skillName, ok := data["skill"].(string); ok {
		skill.Skill = skillName
	}

	if skillType, ok := data["type"].(string); ok {
		skill.Type = skillType
	}

	return skill
}

// convertToSkillLLMAnalysis 转换技能LLM分析
func convertToSkillLLMAnalysis(data map[string]interface{}) *SkillLLMAnalysis {
	if data == nil {
		return nil
	}

	analysis := &SkillLLMAnalysis{}

	if overallMatch, ok := data["overall_match"].(float64); ok {
		analysis.OverallMatch = overallMatch
	}

	if technicalFit, ok := data["technical_fit"].(float64); ok {
		analysis.TechnicalFit = technicalFit
	}

	if learningCurve, ok := data["learning_curve"].(string); ok {
		analysis.LearningCurve = learningCurve
	}

	if strengthAreas, ok := data["strength_areas"].([]interface{}); ok {
		analysis.StrengthAreas = make([]string, 0, len(strengthAreas))
		for _, sa := range strengthAreas {
			if s, ok := sa.(string); ok {
				analysis.StrengthAreas = append(analysis.StrengthAreas, s)
			}
		}
	}

	if gapAreas, ok := data["gap_areas"].([]interface{}); ok {
		analysis.GapAreas = make([]string, 0, len(gapAreas))
		for _, ga := range gapAreas {
			if s, ok := ga.(string); ok {
				analysis.GapAreas = append(analysis.GapAreas, s)
			}
		}
	}

	if recommendations, ok := data["recommendations"].([]interface{}); ok {
		analysis.Recommendations = make([]string, 0, len(recommendations))
		for _, r := range recommendations {
			if s, ok := r.(string); ok {
				analysis.Recommendations = append(analysis.Recommendations, s)
			}
		}
	}

	if analysisDetail, ok := data["analysis_detail"].(string); ok {
		analysis.AnalysisDetail = analysisDetail
	}

	return analysis
}

// convertToSkillItemAnalysis 转换技能项分析
func convertToSkillItemAnalysis(data map[string]interface{}) *SkillItemAnalysis {
	if data == nil {
		return nil
	}

	analysis := &SkillItemAnalysis{}

	if matchLevel, ok := data["match_level"].(string); ok {
		analysis.MatchLevel = matchLevel
	}

	if matchPercentage, ok := data["match_percentage"].(float64); ok {
		analysis.MatchPercentage = matchPercentage
	}

	if proficiencyGap, ok := data["proficiency_gap"].(string); ok {
		analysis.ProficiencyGap = proficiencyGap
	}

	if transferability, ok := data["transferability"].(string); ok {
		analysis.Transferability = transferability
	}

	if learningEffort, ok := data["learning_effort"].(string); ok {
		analysis.LearningEffort = learningEffort
	}

	if matchReason, ok := data["match_reason"].(string); ok {
		analysis.MatchReason = matchReason
	}

	return analysis
}

// convertToMatchedResponsibility 转换匹配的职责
func convertToMatchedResponsibility(data map[string]interface{}) *MatchedResponsibility {
	if data == nil {
		return nil
	}

	resp := &MatchedResponsibility{}

	if jobRespID, ok := data["job_responsibility_id"].(string); ok {
		resp.JobResponsibilityID = jobRespID
	}

	if resumeExpID, ok := data["resume_experience_id"].(string); ok {
		resp.ResumeExperienceID = resumeExpID
	}

	if llmAnalysis, ok := data["llm_analysis"].(map[string]interface{}); ok {
		resp.LLMAnalysis = convertToLLMMatchAnalysis(llmAnalysis)
	}

	if matchScore, ok := data["match_score"].(float64); ok {
		resp.MatchScore = matchScore
	}

	if matchReason, ok := data["match_reason"].(string); ok {
		resp.MatchReason = matchReason
	}

	return resp
}

// convertToJobResponsibility 转换职位职责
func convertToJobResponsibility(data map[string]interface{}) *JobResponsibility {
	if data == nil {
		return nil
	}

	resp := &JobResponsibility{}

	if id, ok := data["id"].(string); ok {
		resp.ID = id
	}

	if jobID, ok := data["job_id"].(string); ok {
		resp.JobID = jobID
	}

	if responsibility, ok := data["responsibility"].(string); ok {
		resp.Responsibility = responsibility
	}

	return resp
}

// convertToLLMMatchAnalysis 转换LLM匹配分析
func convertToLLMMatchAnalysis(data map[string]interface{}) *LLMMatchAnalysis {
	if data == nil {
		return nil
	}

	analysis := &LLMMatchAnalysis{}

	if matchLevel, ok := data["match_level"].(string); ok {
		analysis.MatchLevel = matchLevel
	}

	if matchPercentage, ok := data["match_percentage"].(float64); ok {
		analysis.MatchPercentage = matchPercentage
	}

	if strengthPoints, ok := data["strength_points"].([]interface{}); ok {
		analysis.StrengthPoints = make([]string, 0, len(strengthPoints))
		for _, sp := range strengthPoints {
			if s, ok := sp.(string); ok {
				analysis.StrengthPoints = append(analysis.StrengthPoints, s)
			}
		}
	}

	if weakPoints, ok := data["weak_points"].([]interface{}); ok {
		analysis.WeakPoints = make([]string, 0, len(weakPoints))
		for _, wp := range weakPoints {
			if s, ok := wp.(string); ok {
				analysis.WeakPoints = append(analysis.WeakPoints, s)
			}
		}
	}

	if recommendedActions, ok := data["recommended_actions"].([]interface{}); ok {
		analysis.RecommendedActions = make([]string, 0, len(recommendedActions))
		for _, ra := range recommendedActions {
			if s, ok := ra.(string); ok {
				analysis.RecommendedActions = append(analysis.RecommendedActions, s)
			}
		}
	}

	if analysisDetail, ok := data["analysis_detail"].(string); ok {
		analysis.AnalysisDetail = analysisDetail
	}

	return analysis
}

// convertToYearsMatchInfo 转换年限匹配信息
func convertToYearsMatchInfo(data map[string]interface{}) *YearsMatchInfo {
	if data == nil {
		return nil
	}

	info := &YearsMatchInfo{}

	if requiredYears, ok := data["required_years"].(float64); ok {
		info.RequiredYears = requiredYears
	}

	if actualYears, ok := data["actual_years"].(float64); ok {
		info.ActualYears = actualYears
	}

	if score, ok := data["score"].(float64); ok {
		info.Score = score
	}

	if gap, ok := data["gap"].(float64); ok {
		info.Gap = gap
	}

	if analysis, ok := data["analysis"].(string); ok {
		info.Analysis = analysis
	}

	return info
}

// convertToPositionMatchInfo 转换职位匹配信息
func convertToPositionMatchInfo(data map[string]interface{}) *PositionMatchInfo {
	if data == nil {
		return nil
	}

	info := &PositionMatchInfo{}

	if resumeExpID, ok := data["resume_experience_id"].(string); ok {
		info.ResumeExperienceID = resumeExpID
	}

	if expType, ok := data["experience_type"].(string); ok {
		info.ExperienceType = expType
	}

	if position, ok := data["position"].(string); ok {
		info.Position = position
	}

	if company, ok := data["company"].(string); ok {
		info.Company = company
	}

	if relevance, ok := data["relevance"].(float64); ok {
		info.Relevance = relevance
	}

	if score, ok := data["score"].(float64); ok {
		info.Score = score
	}

	if analysis, ok := data["analysis"].(string); ok {
		info.Analysis = analysis
	}

	return info
}

// convertToIndustryMatchInfo 转换行业匹配信息
func convertToIndustryMatchInfo(data map[string]interface{}) *IndustryMatchInfo {
	if data == nil {
		return nil
	}

	info := &IndustryMatchInfo{}

	if resumeExpID, ok := data["resume_experience_id"].(string); ok {
		info.ResumeExperienceID = resumeExpID
	}

	if company, ok := data["company"].(string); ok {
		info.Company = company
	}

	if industry, ok := data["industry"].(string); ok {
		info.Industry = industry
	}

	if relevance, ok := data["relevance"].(float64); ok {
		info.Relevance = relevance
	}

	if score, ok := data["score"].(float64); ok {
		info.Score = score
	}

	if analysis, ok := data["analysis"].(string); ok {
		info.Analysis = analysis
	}

	return info
}

// convertToCareerProgressionInfo 转换职业发展轨迹信息
func convertToCareerProgressionInfo(data map[string]interface{}) *CareerProgressionInfo {
	if data == nil {
		return nil
	}

	info := &CareerProgressionInfo{}

	if score, ok := data["score"].(float64); ok {
		info.Score = score
	}

	if trend, ok := data["trend"].(string); ok {
		info.Trend = trend
	}

	if analysis, ok := data["analysis"].(string); ok {
		info.Analysis = analysis
	}

	return info
}

// convertToDegreeMatchInfo 转换学历匹配信息
func convertToDegreeMatchInfo(data map[string]interface{}) *DegreeMatchInfo {
	if data == nil {
		return nil
	}

	info := &DegreeMatchInfo{}

	if requiredDegree, ok := data["required_degree"].(string); ok {
		info.RequiredDegree = requiredDegree
	}

	if actualDegree, ok := data["actual_degree"].(string); ok {
		info.ActualDegree = actualDegree
	}

	if score, ok := data["score"].(float64); ok {
		info.Score = score
	}

	if meets, ok := data["meets"].(bool); ok {
		info.Meets = meets
	}

	return info
}

// convertToMajorMatchInfo 转换专业匹配信息
func convertToMajorMatchInfo(data map[string]interface{}) *MajorMatchInfo {
	if data == nil {
		return nil
	}

	info := &MajorMatchInfo{}

	if resumeEduID, ok := data["resume_education_id"].(string); ok {
		info.ResumeEducationID = resumeEduID
	}

	if major, ok := data["major"].(string); ok {
		info.Major = major
	}

	if relevance, ok := data["relevance"].(float64); ok {
		info.Relevance = relevance
	}

	if score, ok := data["score"].(float64); ok {
		info.Score = score
	}

	return info
}

// convertToSchoolMatchInfo 转换学校匹配信息
func convertToSchoolMatchInfo(data map[string]interface{}) *SchoolMatchInfo {
	if data == nil {
		return nil
	}

	info := &SchoolMatchInfo{}

	if resumeEduID, ok := data["resume_education_id"].(string); ok {
		info.ResumeEducationID = resumeEduID
	}

	if school, ok := data["school"].(string); ok {
		info.School = school
	}

	if reputation, ok := data["reputation"].(float64); ok {
		info.Reputation = reputation
	}

	if degree, ok := data["degree"].(string); ok {
		info.Degree = degree
	}

	if major, ok := data["major"].(string); ok {
		info.Major = major
	}

	if gradYear, ok := data["graduation_year"].(float64); ok {
		info.GraduationYear = int(gradYear)
	}

	if score, ok := data["score"].(float64); ok {
		info.Score = score
	}

	if uniTypes, ok := data["university_types"].([]interface{}); ok {
		info.UniversityTypes = make([]consts.UniversityType, 0, len(uniTypes))
		for _, t := range uniTypes {
			if s, ok := t.(string); ok {
				info.UniversityTypes = append(info.UniversityTypes, consts.UniversityType(s))
			}
		}
	}

	if gpa, ok := data["gpa"].(float64); ok {
		info.GPA = &gpa
	}

	if analysis, ok := data["analysis"].(string); ok {
		info.Analysis = analysis
	}

	return info
}

// convertToCompanyMatchInfo 转换公司匹配信息
func convertToCompanyMatchInfo(data map[string]interface{}) *CompanyMatchInfo {
	if data == nil {
		return nil
	}

	info := &CompanyMatchInfo{}

	if resumeExpID, ok := data["resume_experience_id"].(string); ok {
		info.ResumeExperienceID = resumeExpID
	}

	if company, ok := data["company"].(string); ok {
		info.Company = company
	}

	if targetCompany, ok := data["target_company"].(string); ok {
		info.TargetCompany = targetCompany
	}

	if companySize, ok := data["company_size"].(string); ok {
		info.CompanySize = companySize
	}

	if reputation, ok := data["reputation"].(float64); ok {
		info.Reputation = reputation
	}

	if score, ok := data["score"].(float64); ok {
		info.Score = score
	}

	if isExact, ok := data["is_exact"].(bool); ok {
		info.IsExact = isExact
	}

	if analysis, ok := data["analysis"].(string); ok {
		info.Analysis = analysis
	}

	return info
}

// convertToSkillMatchDetail 转换技能匹配详情
func convertToSkillMatchDetail(data map[string]interface{}) *SkillMatchDetail {
	if data == nil {
		return nil
	}

	detail := &SkillMatchDetail{}

	if score, ok := data["score"].(float64); ok {
		detail.Score = score
	}

	// 转换匹配的技能
	if matchedSkills, ok := data["matched_skills"].([]interface{}); ok {
		detail.MatchedSkills = make([]*MatchedSkill, 0, len(matchedSkills))
		for _, ms := range matchedSkills {
			if skillData, ok := ms.(map[string]interface{}); ok {
				skill := convertToMatchedSkill(skillData)
				if skill != nil {
					detail.MatchedSkills = append(detail.MatchedSkills, skill)
				}
			}
		}
	}

	// 转换缺失的技能
	if missingSkills, ok := data["missing_skills"].([]interface{}); ok {
		detail.MissingSkills = make([]*JobSkill, 0, len(missingSkills))
		for _, ms := range missingSkills {
			if skillData, ok := ms.(map[string]interface{}); ok {
				skill := convertToJobSkill(skillData)
				if skill != nil {
					detail.MissingSkills = append(detail.MissingSkills, skill)
				}
			}
		}
	}

	// 转换额外技能
	if extraSkills, ok := data["extra_skills"].([]interface{}); ok {
		detail.ExtraSkills = make([]string, 0, len(extraSkills))
		for _, es := range extraSkills {
			if s, ok := es.(string); ok {
				detail.ExtraSkills = append(detail.ExtraSkills, s)
			}
		}
	}

	// 转换LLM分析
	if llmAnalysis, ok := data["llm_analysis"].(map[string]interface{}); ok {
		detail.LLMAnalysis = convertToSkillLLMAnalysis(llmAnalysis)
	}

	return detail
}

// convertToResponsibilityMatchDetail 转换职责匹配详情
func convertToResponsibilityMatchDetail(data map[string]interface{}) *ResponsibilityMatchDetail {
	if data == nil {
		return nil
	}

	detail := &ResponsibilityMatchDetail{}

	if score, ok := data["score"].(float64); ok {
		detail.Score = score
	}

	// 转换匹配的职责
	if matchedResp, ok := data["matched_responsibilities"].([]interface{}); ok {
		detail.MatchedResponsibilities = make([]*MatchedResponsibility, 0, len(matchedResp))
		for _, mr := range matchedResp {
			if respData, ok := mr.(map[string]interface{}); ok {
				resp := convertToMatchedResponsibility(respData)
				if resp != nil {
					detail.MatchedResponsibilities = append(detail.MatchedResponsibilities, resp)
				}
			}
		}
	}

	// 转换未匹配的职责
	if unmatchedResp, ok := data["unmatched_responsibilities"].([]interface{}); ok {
		detail.UnmatchedResponsibilities = make([]*JobResponsibility, 0, len(unmatchedResp))
		for _, ur := range unmatchedResp {
			if respData, ok := ur.(map[string]interface{}); ok {
				resp := convertToJobResponsibility(respData)
				if resp != nil {
					detail.UnmatchedResponsibilities = append(detail.UnmatchedResponsibilities, resp)
				}
			}
		}
	}

	// 转换相关经验
	if relevantExp, ok := data["relevant_experiences"].([]interface{}); ok {
		detail.RelevantExperiences = make([]string, 0, len(relevantExp))
		for _, re := range relevantExp {
			if s, ok := re.(string); ok {
				detail.RelevantExperiences = append(detail.RelevantExperiences, s)
			}
		}
	}

	return detail
}

// convertToExperienceMatchDetail 转换经验匹配详情
func convertToExperienceMatchDetail(data map[string]interface{}) *ExperienceMatchDetail {
	if data == nil {
		return nil
	}

	detail := &ExperienceMatchDetail{}

	if score, ok := data["score"].(float64); ok {
		detail.Score = score
	}

	// 转换年限匹配信息
	if yearsMatch, ok := data["years_match"].(map[string]interface{}); ok {
		detail.YearsMatch = convertToYearsMatchInfo(yearsMatch)
	}

	// 转换职位匹配信息
	if positionMatches, ok := data["position_matches"].([]interface{}); ok {
		detail.PositionMatches = make([]*PositionMatchInfo, 0, len(positionMatches))
		for _, pm := range positionMatches {
			if posData, ok := pm.(map[string]interface{}); ok {
				pos := convertToPositionMatchInfo(posData)
				if pos != nil {
					detail.PositionMatches = append(detail.PositionMatches, pos)
				}
			}
		}
	}

	// 转换行业匹配信息
	if industryMatches, ok := data["industry_matches"].([]interface{}); ok {
		detail.IndustryMatches = make([]*IndustryMatchInfo, 0, len(industryMatches))
		for _, im := range industryMatches {
			if indData, ok := im.(map[string]interface{}); ok {
				ind := convertToIndustryMatchInfo(indData)
				if ind != nil {
					detail.IndustryMatches = append(detail.IndustryMatches, ind)
				}
			}
		}
	}

	// 转换职业发展轨迹
	if progression, ok := data["career_progression"].(map[string]interface{}); ok {
		detail.CareerProgression = convertToCareerProgressionInfo(progression)
	}

	// 转换整体分析
	if overall, ok := data["overall_analysis"].(string); ok {
		detail.OverallAnalysis = overall
	}

	return detail
}

// convertToEducationMatchDetail 转换教育匹配详情
func convertToEducationMatchDetail(data map[string]interface{}) *EducationMatchDetail {
	if data == nil {
		return nil
	}

	detail := &EducationMatchDetail{}

	if score, ok := data["score"].(float64); ok {
		detail.Score = score
	}

	// 转换学历匹配信息
	if degreeMatch, ok := data["degree_match"].(map[string]interface{}); ok {
		detail.DegreeMatch = convertToDegreeMatchInfo(degreeMatch)
	}

	// 转换专业匹配信息
	if majorMatches, ok := data["major_matches"].([]interface{}); ok {
		detail.MajorMatches = make([]*MajorMatchInfo, 0, len(majorMatches))
		for _, mm := range majorMatches {
			if majorData, ok := mm.(map[string]interface{}); ok {
				major := convertToMajorMatchInfo(majorData)
				if major != nil {
					detail.MajorMatches = append(detail.MajorMatches, major)
				}
			}
		}
	}

	// 转换学校匹配信息
	if schoolMatches, ok := data["school_matches"].([]interface{}); ok {
		detail.SchoolMatches = make([]*SchoolMatchInfo, 0, len(schoolMatches))
		for _, sm := range schoolMatches {
			if schoolData, ok := sm.(map[string]interface{}); ok {
				school := convertToSchoolMatchInfo(schoolData)
				if school != nil {
					detail.SchoolMatches = append(detail.SchoolMatches, school)
				}
			}
		}
	}

	// 转换整体分析
	if overall, ok := data["overall_analysis"].(string); ok {
		detail.OverallAnalysis = overall
	}

	return detail
}

// convertToIndustryMatchDetail 转换行业匹配详情
func convertToIndustryMatchDetail(data map[string]interface{}) *IndustryMatchDetail {
	if data == nil {
		return nil
	}

	detail := &IndustryMatchDetail{}

	if score, ok := data["score"].(float64); ok {
		detail.Score = score
	}

	// 转换行业匹配信息
	if industryMatches, ok := data["industry_matches"].([]interface{}); ok {
		detail.IndustryMatches = make([]*IndustryMatchInfo, 0, len(industryMatches))
		for _, im := range industryMatches {
			if indData, ok := im.(map[string]interface{}); ok {
				ind := convertToIndustryMatchInfo(indData)
				if ind != nil {
					detail.IndustryMatches = append(detail.IndustryMatches, ind)
				}
			}
		}
	}

	// 转换公司匹配信息
	if companyMatches, ok := data["company_matches"].([]interface{}); ok {
		detail.CompanyMatches = make([]*CompanyMatchInfo, 0, len(companyMatches))
		for _, cm := range companyMatches {
			if compData, ok := cm.(map[string]interface{}); ok {
				comp := convertToCompanyMatchInfo(compData)
				if comp != nil {
					detail.CompanyMatches = append(detail.CompanyMatches, comp)
				}
			}
		}
	}

	// 转换行业深度评估
	if depthData, ok := data["industry_depth"].(map[string]interface{}); ok {
		depth := &IndustryDepthInfo{}
		if years, ok := depthData["total_years"].(float64); ok {
			depth.TotalYears = years
		}
		if dscore, ok := depthData["score"].(float64); ok {
			depth.Score = dscore
		}
		if analysis, ok := depthData["analysis"].(string); ok {
			depth.Analysis = analysis
		}
		detail.IndustryDepth = depth
	}

	// 转换整体分析
	if overall, ok := data["overall_analysis"].(string); ok {
		detail.OverallAnalysis = overall
	}

	return detail
}
