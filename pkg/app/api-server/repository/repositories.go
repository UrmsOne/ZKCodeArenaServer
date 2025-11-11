/*
@Author: omenkk7
@Date: 2025/10/25
@Description: Repository 容器 - 管理所有Repository实例
*/

package repository

type Repositories struct {
	ProblemRepository  *ProblemRepository
	SubmitRepository   *SubmitRepository
	TestCaseRepository *TestCaseRepository
}

// NewRepositories 创建Repository容器实例
func NewRepositories() *Repositories {
	return &Repositories{
		ProblemRepository:  NewProblemRepository(),
		SubmitRepository:   NewSubmitRepository(),
		TestCaseRepository: NewTestCaseRepository(),
	}
}
