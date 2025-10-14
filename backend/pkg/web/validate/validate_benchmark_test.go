package validate

import (
	"reflect"
	"testing"
)

// BenchmarkSimpleStruct 测试简单结构体的性能
func BenchmarkSimpleStruct(b *testing.B) {
	validator := NewCustomValidator()
	userReq := &UserRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(userReq)
	}
}

// BenchmarkNestedStruct 测试嵌套结构体的性能
func BenchmarkNestedStruct(b *testing.B) {
	validator := NewCustomValidator()
	productReq := &ProductRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(productReq)
	}
}

// BenchmarkStructWithValues 测试有值的结构体性能
func BenchmarkStructWithValues(b *testing.B) {
	validator := NewCustomValidator()
	userReq := &UserRequest{
		Name:  "测试用户",
		Age:   25,
		Email: "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(userReq)
	}
}

// BenchmarkNestedStructWithValues 测试有值的嵌套结构体性能
func BenchmarkNestedStructWithValues(b *testing.B) {
	validator := NewCustomValidator()
	productReq := &ProductRequest{
		Name:  "测试产品",
		Price: 99.99,
		Address: Address{
			Street: "测试街道",
			City:   "测试城市",
		},
		Contact: &ContactInfo{
			Phone: "123456789",
			Email: "contact@example.com",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(productReq)
	}
}

// BenchmarkApplyDefaultsOnly 只测试应用默认值的性能
func BenchmarkApplyDefaultsOnly(b *testing.B) {
	validator := NewCustomValidator()
	userReq := &UserRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.applyDefaults(userReq)
	}
}

// BenchmarkApplyDefaultsNested 测试嵌套结构体应用默认值的性能
func BenchmarkApplyDefaultsNested(b *testing.B) {
	validator := NewCustomValidator()
	productReq := &ProductRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.applyDefaults(productReq)
	}
}

// BenchmarkSetDefault 测试手动设置默认值的性能
func BenchmarkSetDefault(b *testing.B) {
	validator := NewCustomValidator()
	userReq := &UserRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.SetDefault(userReq, "Name", "默认名称")
	}
}

// BenchmarkValidatorCreation 测试验证器创建的性能
func BenchmarkValidatorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewCustomValidator()
	}
}

// BenchmarkIsZeroValue 测试零值检查的性能
func BenchmarkIsZeroValue(b *testing.B) {
	validator := NewCustomValidator()
	v := reflect.ValueOf("")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.isZeroValue(v)
	}
}

// BenchmarkSetFieldValue 测试字段值设置的性能
func BenchmarkSetFieldValue(b *testing.B) {
	validator := NewCustomValidator()
	userReq := &UserRequest{}
	field := reflect.ValueOf(userReq).Elem().FieldByName("Name")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.setFieldValue(field, "测试名称")
	}
}

// 创建更复杂的嵌套结构体用于性能测试
type ComplexAddress struct {
	Street     string `json:"street" default:"未知街道"`
	City       string `json:"city" default:"未知城市"`
	Province   string `json:"province" default:"未知省份"`
	ZipCode    string `json:"zip_code" default:"000000"`
	Country    string `json:"country" default:"中国"`
	PostalCode string `json:"postal_code" default:"000000"`
}

type ComplexContact struct {
	Phone   string `json:"phone" default:"无"`
	Email   string `json:"email" validate:"email" default:"contact@example.com"`
	Website string `json:"website" default:"https://example.com"`
	Fax     string `json:"fax" default:"无"`
	Mobile  string `json:"mobile" default:"无"`
	WeChat  string `json:"wechat" default:"无"`
}

type ComplexProduct struct {
	Name        string          `json:"name" validate:"required"`
	Price       float64         `json:"price" validate:"gte=0" default:"0.0"`
	Description string          `json:"description" default:"暂无描述"`
	Category    string          `json:"category" default:"未分类"`
	InStock     bool            `json:"in_stock" default:"false"`
	Address     ComplexAddress  `json:"address"`
	Contact     *ComplexContact `json:"contact"`
	Tags        []string        `json:"tags" default:"[]"`
	Rating      float64         `json:"rating" validate:"gte=0,lte=5" default:"0.0"`
	Reviews     int             `json:"reviews" validate:"gte=0" default:"0"`
}

// BenchmarkComplexNestedStruct 测试复杂嵌套结构体的性能
func BenchmarkComplexNestedStruct(b *testing.B) {
	validator := NewCustomValidator()
	productReq := &ComplexProduct{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(productReq)
	}
}

// BenchmarkComplexNestedStructWithValues 测试有值的复杂嵌套结构体性能
func BenchmarkComplexNestedStructWithValues(b *testing.B) {
	validator := NewCustomValidator()
	productReq := &ComplexProduct{
		Name:  "复杂产品",
		Price: 199.99,
		Address: ComplexAddress{
			Street: "复杂街道",
			City:   "复杂城市",
		},
		Contact: &ComplexContact{
			Phone: "123456789",
			Email: "complex@example.com",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(productReq)
	}
}
