package model

// 难度等级
type Level struct {
    Min  int    `json:"min"`
    Max  int    `json:"max"`
    Name string `json:"name"`
}

// 用户界面语言
type UserLang struct {
    ID          string `json:"id"`
    Description string `json:"description"`
    Flag        string `json:"flag,omitempty"`
}

// 编程语言
type CodeLang struct {
    ID          string  `json:"id"`
    Ext         []string `json:"ext"`
    Description string  `json:"description"`
    Ratio       float64 `json:"ratio"`
}

// 判题结果
type Verdict struct {
    ID          string `json:"id"`
    Description string `json:"description"`
    Abbr        string `json:"abbr"`
    Color       string `json:"color"`
}

// 标签分类
type TagClassify struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// 标签
type Tag struct {
    ID       int    `json:"id"`
    Classify int    `json:"classify"`
    Name     string `json:"name"`
}

// 同步配置响应
type SyncConfigResponse struct {
    Tags        map[int]Tag         `json:"tags"`
    UserLangs   map[string]UserLang `json:"userLangs"`
    CodeLangs   map[string]CodeLang `json:"codeLangs"`
    Verdicts    map[string]Verdict  `json:"verdicts"`
    Difficulties []Level            `json:"difficulties"`
}