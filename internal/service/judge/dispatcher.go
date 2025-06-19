package judge

import (
	"context"
	"os"
	"path/filepath"
	"reisen-be/internal/filesystem"
	"reisen-be/internal/model"
	"reisen-be/internal/websocket"
	"sync"
	"time"
)

// 每个 Dispatcher 管理一个评测记录从提交评测、编译、运行、判分的全过程
type Dispatcher struct {
	compiler          *Compiler
	runner            *Runner
	checker           Checker
	taskQueue         chan *model.JudgeTask
	submissionChan    chan *model.Submission
	workers           int
	problemFilesystem *filesystem.ProblemFilesystem
	submissionWs      *websocket.SubmissionWs
}
func NewDispatcher(workers int, compiler *Compiler, runner *Runner, checker Checker, problemFilesystem *filesystem.ProblemFilesystem, submissionWs *websocket.SubmissionWs) *Dispatcher {
	return &Dispatcher{
		compiler:          compiler,
		runner:            runner,
		checker:           checker,
		taskQueue:         make(chan *model.JudgeTask, 100),
		submissionChan:    make(chan *model.Submission, 100),
		workers:           workers,
		problemFilesystem: problemFilesystem,
		submissionWs:      submissionWs,
	}
}

// 创建 worker 实例
func (d *Dispatcher) Start(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < d.workers; i++ {
		wg.Add(1)
		go d.worker(ctx, &wg)
	}

	go func() {
		wg.Wait()
		close(d.submissionChan)
	}()
}

// 提交评测任务
func (d *Dispatcher) Submit(task *model.JudgeTask) {
	d.taskQueue <- task
}

// 收集评测结果（Submission）
func (d *Dispatcher) Results() <- chan *model.Submission {
	return d.submissionChan
}

func (d *Dispatcher) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// 从评测队列获取评测任务
	for {
		select {
		case task := <-d.taskQueue:
			d.judgeTask(task)
			d.submissionChan <- &task.Submission
		case <-ctx.Done():
			return
		}
	}
}

// 运行任务并写入运行结果
func (d *Dispatcher) judgeTask(task *model.JudgeTask) {

	task.Verdict = model.VerdictJD

	// 1. 编译代码
	fileId, compileInfo, err := d.compiler.Compile(task)
	task.CompileInfo = compileInfo

	if err != nil {
		task.Verdict = model.VerdictCE
		for i := range task.Testcases {
			task.Testcases[i].Verdict = model.VerdictCE
		}
		d.submissionWs.Broadcast(task.ID, task.Submission)
		return
	}
	defer d.compiler.DeleteFile(fileId)

	// 2. 运行测试用例
	var wg sync.WaitGroup
	testCaseChan := make(chan int, len(task.Config.TestCases)) // 用于通知完成的测试点索引

	// 编译通过，生成评测信息
	for i, tc := range task.Config.TestCases {
		wg.Add(1)
		go func(idx int, testCase model.TestCaseConfig) {
			defer func() {
				testCaseChan <- idx
			}()
			defer wg.Done()

			testResult := &task.Testcases[idx]

			runResult, err := d.runner.Run(task, fileId, testCase, d.problemFilesystem.GetProblemPath(task.ProblemID))
			if err != nil {
				message := err.Error()
				testResult.Verdict = model.VerdictUKE
				testResult.Checker = &message
				return
			} else {
				*testResult = *runResult
				
				if testResult.Verdict != model.VerdictAC {
					return
				}
			}

			// 3. 判分

			// 提取答案文件
			expectedOutput, err := os.ReadFile(filepath.Join(d.problemFilesystem.GetProblemPath(task.ProblemID), testCase.OutputFile))
			if err != nil {
				message := err.Error()
				testResult.Verdict = model.VerdictUKE
				testResult.Checker = &message
				return
			}

			// 检查是否通过
			passed, message := d.checker.Check(*testResult.Output, string(expectedOutput))
			if message != "" {
				// 校验器输出信息
				testResult.Checker = &message
			}
			if passed {
				testResult.Verdict = model.VerdictAC
				testResult.Score = &testCase.Score
			} else {
				testResult.Verdict = model.VerdictWA
			}

		}(i, tc)
	}
	
	// 启动协程监听测试点完成
	go func() {
		lastUpdate := time.Now()
		updatedCount := 0
		
		for range testCaseChan {
			updatedCount ++
			
			// 满足以下条件之一时更新评测结果：
			// 1. 已有 10 个以上测试点状态被更新
			// 2. 已有  5 秒以上测试点状态未更新
			if updatedCount >= 10 || time.Since(lastUpdate) > time.Second * 5 {
				d.submissionWs.Broadcast(task.ID, task.Submission)
				lastUpdate = time.Now()
				updatedCount = 0
			}
		}
	}()

	wg.Wait()
	// 关闭 testCaseChan，通知监听 goroutine 退出
	close(testCaseChan)

	// 等待评测完毕，收集结果
	var maxTimeUsed, maxMemoryUsed, totalScore int
	allPassed := true

	for i := range task.Testcases {
		tr := &task.Testcases[i]
		if tr.Time != nil {
			maxTimeUsed = max(*tr.Time, maxTimeUsed)
		}
		if tr.Memory != nil {
			maxMemoryUsed = max(*tr.Memory, maxMemoryUsed)
		}
		if tr.Score != nil {
			totalScore += *tr.Score
		}
		if tr.Input != nil && len(*tr.Input) > 256 {
			*tr.Input = (*tr.Input)[0:256] + "..."
		}
		if tr.Output != nil && len(*tr.Output) > 256 {
				*tr.Output = (*tr.Output)[0:256] + "..."
		}
		if allPassed && tr.Verdict != model.VerdictAC {
			allPassed = false
			task.Verdict = tr.Verdict
		}
	}

	// 全部通过
	if allPassed {
		task.Verdict = model.VerdictAC
	}

	task.TimeUsed = &maxTimeUsed
	task.MemoryUsed = &maxMemoryUsed
	task.Score = &totalScore

	// 未知错误
	if task.Verdict == "JD" || task.Verdict == "PD" {
		task.Verdict = model.VerdictUKE
	}
	
	// 广播评测结果
	d.submissionWs.Broadcast(task.ID, task.Submission)
}
