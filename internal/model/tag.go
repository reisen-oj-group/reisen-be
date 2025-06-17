package model

// TagClassify 标签分类
type TagClassify struct {
	ID   TagClassifyId `gorm:"primaryKey"`
	Name string
}

func (TagClassify) TableName() string {
	return "tag_classifies"
}

// Tag 标签
type Tag struct {
	ID       TagId `gorm:"primaryKey"`
	Name     string
	Classify TagClassifyId
}

func (Tag) TableName() string {
	return "tags"
}

// 标签请求
type TagRequest struct {
	Tag TagId `json:"tag"`
}

// 标签响应
type TagResponse struct {
	Tag Tag `json:"tag"`
}

// 标签列表请求
type TagListRequest struct {
	Classify *TagClassifyId `json:"omitempty"`
}

// 标签列表响应
type TagListResponse struct {
	Classifies []TagClassify
	Tags       []Tag
}

// 标签编辑请求
type TagEditRequest struct {
	Tag Tag
}

// 标签编辑响应
type TagEditResponse struct {
	Tag Tag
}

// 标签删除请求
type TagDeleteRequest struct {
	Tag TagId
}

// 标签删除响应
type TagDeleteResponse struct {
}
