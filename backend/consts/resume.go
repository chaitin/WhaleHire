package consts

type UniversityType string

const (
	UniversityTypeOrdinary UniversityType = "ordinary" // 普通学校
	UniversityType211      UniversityType = "211"      // 211工程学校
	UniversityType985      UniversityType = "985"      // 985工程学校
)

const (
	// 批量上传相关Redis键值
	BatchUploadTaskKeyFmt = "batch_upload:task:%s"
	BatchUploadItemKeyFmt = "batch_upload:item:%s:%s" // taskID:itemID
)
