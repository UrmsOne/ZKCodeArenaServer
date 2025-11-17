# 实现计划

- [x] 1. 在 BaseRepository 中添加批量操作方法


  - 实现 `DeleteMany` 方法支持批量删除文档
  - 实现 `UpdateMany` 方法支持批量更新文档，自动添加 updated_at 时间戳
  - 添加适当的日志记录和错误处理
  - _需求: 1.1, 1.2, 1.3, 1.4_


- [ ] 2. 集成 pkg/common/mongo/query.go 的查询功能
  - 在 BaseRepository 中导入 `common "zk-code-arena-server/pkg/common/mongo"` 包
  - 实现 `FindWithPaginationTyped[T any]` 泛型分页查询方法，直接调用 `common.FindWithPagination`
  - 实现 `BuildSort` 方法，直接转发到 `common.BuildSort`
  - 实现 `BuildKeywordFilter` 方法，直接转发到 `common.BuildKeywordFilter`
  - 实现 `MergeFilters` 方法，直接转发到 `common.MergeFilters`
  - 实现 `BuildFindOptionsWithPagination` 方法，直接转发到 `common.BuildFindOptions`
  - _需求: 2.1, 2.2, 2.3, 2.4, 3.1, 3.2, 3.3, 3.4, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4_



- [ ] 3. 更新 ProblemRepository 使用新的查询方法
  - 重构 `GetProblems` 方法使用 `FindWithPaginationTyped` 和 `BuildFindOptionsWithPagination`
  - 重构 `GetProblemsWithUserStatus` 方法使用新的查询方法
  - 重构 `SearchProblems` 方法使用 `BuildKeywordFilter` 和 `MergeFilters`
  - 简化查询逻辑，移除冗余的类型转换代码
  - _需求: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3_

- [x] 4. 更新 SubmitRepository 使用新的查询方法





  - 重构 `GetSubmits` 方法使用 `FindWithPaginationTyped` 替代旧的 `FindWithPagination`
  - 重构 `GetSubmitsByStatus` 方法使用新的泛型查询方法
  - 重构 `GetSubmitsList` 方法使用新的查询方法
  - 重构 `BatchUpdateStatus` 方法使用 BaseRepository 的 `UpdateMany` 方法
  - 简化查询逻辑，移除冗余的类型转换代码
  - _需求: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3_




- [x] 5. 更新 TestcaseRepository 使用新的查询方法






  - 重构 `GetByProblemID` 方法使用 `FindWithPaginationTyped` 替代旧的 `FindWithPagination`
  - 重构 `GetSampleTestCases` 方法使用新的泛型查询方法
  - 重构 `DeleteByProblemID` 方法使用 BaseRepository 的 `DeleteMany` 方法
  - 简化查询逻辑，移除冗余的类型转换代码
  - _需求: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3_

- [ ]* 6. 验证功能正确性
  - 测试批量删除功能是否正常工作
  - 测试批量更新功能是否正常工作并自动更新时间戳
  - 测试分页查询功能返回正确的数据和总数
  - 测试排序功能（升序、降序、多字段排序）
  - 测试关键字搜索功能（单字段、多字段、特殊字符）
  - 测试过滤器合并功能
  - 验证 ProblemRepository 的查询方法返回正确结果
  - 验证 SubmitRepository 的查询方法返回正确结果
  - 验证 TestcaseRepository 的查询方法返回正确结果
  - _需求: 所有需求_
