package validate

import (
	"fmt"
	"log"
)

// UserRequest 用户请求结构体示例
type UserRequest struct {
	Name     string `json:"name" validate:"required" default:"匿名用户"`
	Age      int    `json:"age" validate:"gte=0,lte=150" default:"18"`
	Email    string `json:"email" validate:"email" default:"user@example.com"`
	IsActive bool   `json:"is_active" default:"true"`
	Page     int    `json:"page" validate:"gte=1" default:"1"`
	Size     int    `json:"size" validate:"gte=1,lte=100" default:"10"`
}

// Address 地址结构体
type Address struct {
	Street   string `json:"street" default:"未知街道"`
	City     string `json:"city" default:"未知城市"`
	Province string `json:"province" default:"未知省份"`
	ZipCode  string `json:"zip_code" default:"000000"`
}

// ContactInfo 联系信息结构体
type ContactInfo struct {
	Phone   string `json:"phone" default:"无"`
	Email   string `json:"email" validate:"email" default:"contact@example.com"`
	Website string `json:"website" default:"https://example.com"`
}

// ProductRequest 产品请求结构体示例
type ProductRequest struct {
	Name        string       `json:"name" validate:"required"`
	Price       float64      `json:"price" validate:"gte=0" default:"0.0"`
	Description string       `json:"description" default:"暂无描述"`
	Category    string       `json:"category" default:"未分类"`
	InStock     bool         `json:"in_stock" default:"false"`
	Address     Address      `json:"address"` // 嵌套结构体
	Contact     *ContactInfo `json:"contact"` // 嵌套结构体指针
}

// ExampleUsage 展示如何使用支持默认值的验证器
func ExampleUsage() {
	// 创建验证器
	validator := NewCustomValidator()

	// 示例1: 使用结构体标签设置默认值
	fmt.Println("=== 示例1: 结构体标签默认值 ===")

	// 创建一个空的用户请求（模拟从JSON解析后的空字段）
	userReq := &UserRequest{}

	// 验证并应用默认值
	if err := validator.Validate(userReq); err != nil {
		log.Printf("验证失败: %v", err)
	} else {
		fmt.Printf("用户请求: %+v\n", userReq)
	}

	// 示例2: 部分字段有值，部分字段使用默认值
	fmt.Println("\n=== 示例2: 部分字段有值 ===")

	userReq2 := &UserRequest{
		Name: "张三",
		Age:  25,
		// Email, IsActive, Page, Size 将使用默认值
	}

	if err := validator.Validate(userReq2); err != nil {
		log.Printf("验证失败: %v", err)
	} else {
		fmt.Printf("用户请求2: %+v\n", userReq2)
	}

	// 示例3: 嵌套结构体默认值
	fmt.Println("\n=== 示例3: 嵌套结构体默认值 ===")

	productReq := &ProductRequest{
		Name: "iPhone 15",
		// Address 和 Contact 将使用默认值
	}

	if err := validator.Validate(productReq); err != nil {
		log.Printf("验证失败: %v", err)
	} else {
		fmt.Printf("产品请求: %+v\n", productReq)
		fmt.Printf("地址信息: %+v\n", productReq.Address)
		fmt.Printf("联系信息: %+v\n", productReq.Contact)
	}

	// 示例4: 手动设置默认值
	fmt.Println("\n=== 示例4: 手动设置默认值 ===")

	productReq2 := &ProductRequest{
		Name: "MacBook Pro",
	}

	// 手动设置价格默认值
	if err := validator.SetDefault(productReq2, "Price", "1999.99"); err != nil {
		log.Printf("设置默认值失败: %v", err)
	}

	if err := validator.Validate(productReq2); err != nil {
		log.Printf("验证失败: %v", err)
	} else {
		fmt.Printf("产品请求2: %+v\n", productReq2)
	}

	// 示例5: 验证失败的情况
	fmt.Println("\n=== 示例5: 验证失败 ===")

	invalidUser := &UserRequest{
		Name: "李四",
		Age:  200, // 超出验证范围
	}

	if err := validator.Validate(invalidUser); err != nil {
		fmt.Printf("验证失败: %v\n", err)
	} else {
		fmt.Printf("用户请求: %+v\n", invalidUser)
	}
}

// TestValidator 测试验证器功能
func TestValidator() {
	fmt.Println("开始测试支持默认值的验证器...")
	ExampleUsage()
	fmt.Println("测试完成！")
}
