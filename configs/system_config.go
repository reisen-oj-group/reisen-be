package configs

import "reisen-be/internal/model"

var SystemConfig = model.SyncConfigResponse{
    Tags: map[int]model.Tag{
        1: {ID: 1, Name: "动态规划", Classify: 1},
        2: {ID: 2, Name: "图论", Classify: 1},
        3: {ID: 3, Name: "数据结构", Classify: 1},
    },
    UserLangs: map[string]model.UserLang{
        "en-US": {ID: "en-US", Description: "English"},
        "zh-CN": {ID: "zh-CN", Description: "简体中文"},
        "zh-TW": {ID: "zh-TW", Description: "繁体中文"},
    },
    CodeLangs: map[string]model.CodeLang{
        "cpp":    {ID: "cpp", Ext: []string{".cpp", ".cc", ".cxx"}, Description: "C++ 11", Ratio: 1},
        "java":   {ID: "java", Ext: []string{".java"}, Description: "Java 11", Ratio: 2},
        "python": {ID: "python", Ext: []string{".py"}, Description: "Python 3.8", Ratio: 3},
    },
    Verdicts: map[string]model.Verdict{
        "AC":  {ID: "AC", Description: "Accepted", Abbr: "AC", Color: "#67C23A"},
        "WA":  {ID: "WA", Description: "Wrong Answer", Abbr: "WA", Color: "#F56C6C"},
        "RE":  {ID: "RE", Description: "Runtime Error", Abbr: "RE", Color: "#6A3BC0"},
        "TLE": {ID: "TLE", Description: "Time Limit Exceeded", Abbr: "TLE", Color: "#E6A23C"},
        "MLE": {ID: "MLE", Description: "Memory Limit Exceeded", Abbr: "MLE", Color: "#E6A23C"},
        "CE":  {ID: "CE", Description: "Compile Error", Abbr: "CE", Color: "#909399"},
        "UKE": {ID: "UKE", Description: "Unknown Error", Abbr: "UKE", Color: "#909399"},
    },
    Difficulties: []model.Level{
        {Min: 800, Max: 1099, Name: "入门"},
        {Min: 1100, Max: 1399, Name: "简单"},
        {Min: 1400, Max: 1699, Name: "中等"},
        {Min: 1700, Max: 1999, Name: "较难"},
        {Min: 2000, Max: 2299, Name: "困难"},
        {Min: 2300, Max: 2599, Name: "挑战"},
        {Min: 2600, Max: 2899, Name: "精英"},
        {Min: 2900, Max: 3199, Name: "专家"},
        {Min: 3200, Max: 3500, Name: "大师"},
    },
}