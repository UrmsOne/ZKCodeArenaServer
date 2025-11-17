# 需求文档

## 简介

为 BaseRepository 添加通用的分页查询功能，整合 `pkg/common/mongo/query.go` 中的分页、排序、关键字搜索等功能，使得 repository 层的所有实现都能基于这些通用方法完成数据查询操作。

## 术语表

- **BaseRepository**: 基础仓储类，提供通用的数据访问方法
- **Pagination**: 分页功能，包含页码、每页大小、总数等信息
- **QueryHelper**: 查询辅助模块，位于 `pkg/common/mongo/query.go`
- **Repository Layer**: 仓储层，负责数据访问的抽象层
- **MongoDB Collection**: MongoDB 集合，数据存储的基本单位

## 需求

### 需求 1

**用户故事:** 作为开发者，我希望 BaseRepository 提供统一的分页查询方法，以便在所有 repository 实现中复用分页逻辑

#### 验收标准

1. WHEN 调用分页查询方法时，THE BaseRepository SHALL 使用 QueryHelper 中的 `ComputeSkipLimit` 函数计算跳过和限制参数
2. WHEN 提供页码和每页大小参数时，THE BaseRepository SHALL 自动标准化这些参数（页码最小为1，每页大小在默认值和最大值之间）
3. WHEN 执行分页查询时，THE BaseRepository SHALL 返回数据列表和总记录数
4. THE BaseRepository SHALL 提供泛型方法 `FindWithPagination` 支持任意类型的数据查询

### 需求 2

**用户故事:** 作为开发者，我希望 BaseRepository 支持灵活的排序功能，以便根据不同字段对查询结果进行排序

#### 验收标准

1. WHEN 提供排序表达式时，THE BaseRepository SHALL 使用 QueryHelper 的 `BuildSort` 函数解析排序规则
2. THE BaseRepository SHALL 支持多字段排序，字段之间用逗号分隔
3. THE BaseRepository SHALL 支持升序和降序排序（字段前加 `-` 表示降序）
4. WHEN 排序表达式为空时，THE BaseRepository SHALL 不应用任何排序

### 需求 3

**用户故事:** 作为开发者，我希望 BaseRepository 支持关键字搜索功能，以便在多个字段中进行模糊匹配查询

#### 验收标准

1. WHEN 提供关键字和搜索字段列表时，THE BaseRepository SHALL 使用 QueryHelper 的 `BuildKeywordFilter` 函数构建搜索过滤器
2. THE BaseRepository SHALL 对关键字进行正则表达式转义，防止注入攻击
3. THE BaseRepository SHALL 支持不区分大小写的模糊匹配
4. WHEN 关键字为空或字段列表为空时，THE BaseRepository SHALL 返回空过滤器

### 需求 4

**用户故事:** 作为开发者，我希望 BaseRepository 提供过滤器合并功能，以便组合多个查询条件

#### 验收标准

1. WHEN 提供多个过滤器时，THE BaseRepository SHALL 使用 QueryHelper 的 `MergeFilters` 函数合并所有非空过滤器
2. WHEN 只有一个非空过滤器时，THE BaseRepository SHALL 直接返回该过滤器
3. WHEN 有多个非空过滤器时，THE BaseRepository SHALL 使用 `$and` 操作符组合它们
4. WHEN 所有过滤器都为空时，THE BaseRepository SHALL 返回空过滤器

### 需求 5

**用户故事:** 作为开发者，我希望 BaseRepository 提供构建查询选项的方法，以便统一管理分页、排序和投影

#### 验收标准

1. THE BaseRepository SHALL 提供 `BuildFindOptions` 方法接受页码、每页大小、排序和投影参数
2. WHEN 构建查询选项时，THE BaseRepository SHALL 自动计算 skip 和 limit 值
3. WHEN 提供排序规则时，THE BaseRepository SHALL 将其应用到查询选项中
4. WHEN 提供投影参数时，THE BaseRepository SHALL 将其应用到查询选项中以限制返回字段

### 需求 6

**用户故事:** 作为开发者，我希望 BaseRepository 提供 URL 查询参数解析功能，以便从 HTTP 请求中提取分页和搜索参数

#### 验收标准

1. THE BaseRepository SHALL 提供 `ParseCommonQueryParams` 方法解析 URL 查询参数
2. THE BaseRepository SHALL 从查询参数中提取 `page`、`pageSize`、`sort` 和 `q`（关键字）参数
3. WHEN 参数缺失或无效时，THE BaseRepository SHALL 使用默认值
4. THE BaseRepository SHALL 返回标准化后的参数值
