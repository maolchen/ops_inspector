package model

import (
	"fmt"
	"ops-inspection/internal/config"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite" // 纯 Go 实现，无需 CGO
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(cfg *config.DatabaseConfig) error {
	var err error

	// 配置 GORM 日志
	logLevel := logger.Info
	if config.GlobalConfig.Server.Mode == "release" {
		logLevel = logger.Warn
	}

	// 使用配置文件中的数据库路径，支持持久化
	dsn := cfg.Path
	if dsn == "" {
		dsn = "./data/inspection.db"
	}

	// 确保数据目录存在
	dbDir := filepath.Dir(dsn)
	if dbDir != "" && dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移
	if err = autoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化默认数据
	if err = initDefaultData(); err != nil {
		return fmt.Errorf("failed to init default data: %w", err)
	}

	return nil
}

// autoMigrate 自动迁移表结构
func autoMigrate() error {
	return DB.AutoMigrate(
		&User{},
		&Project{},
		&RuleGroup{},
		&Rule{},
		&InspectionReport{},
		&InspectionItem{},
		&SystemConfig{},
	)
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	// 初始化默认用户
	if err := initDefaultUser(); err != nil {
		return err
	}

	// 初始化默认规则组
	if err := initDefaultRuleGroups(); err != nil {
		return err
	}

	// 初始化默认规则
	if err := initDefaultRules(); err != nil {
		return err
	}

	// 初始化系统配置
	if err := InitDefaultConfigs(); err != nil {
		return err
	}

	return nil
}

// initDefaultUser 初始化默认用户
func initDefaultUser() error {
	var count int64
	DB.Model(&User{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 使用 bcrypt 动态生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 默认用户 admin/admin
	defaultUser := User{
		Username:    "admin",
		Password:    string(hashedPassword),
		DisplayName: "系统管理员",
	}

	return DB.Create(&defaultUser).Error
}

// initDefaultRuleGroups 初始化默认规则组
func initDefaultRuleGroups() error {
	var count int64
	DB.Model(&RuleGroup{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 创建默认规则组（保持code一致）
	defaultGroups := []RuleGroup{
		{Name: "服务器基础资源", Code: "basic_resources", Description: "服务器CPU、内存、磁盘、网络等基础监控", SortOrder: 1},
		{Name: "Kubernetes集群状态", Code: "k8s_cluster", Description: "K8s节点、Pod、PVC等集群状态监控", SortOrder: 2},
		{Name: "进程指标", Code: "process_metrics", Description: "进程CPU、内存使用率监控", SortOrder: 3},
		{Name: "其他指标", Code: "other_metrics", Description: "域名证书等其他监控指标", SortOrder: 4},
	}

	return DB.Create(&defaultGroups).Error
}

// initDefaultRules 初始化默认规则
func initDefaultRules() error {
	var count int64
	DB.Model(&Rule{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 获取规则组ID映射
	var groups []RuleGroup
	DB.Find(&groups)
	groupMap := make(map[string]uint)
	for _, g := range groups {
		groupMap[g.Code] = g.ID
	}

	// 兼容映射：处理不同版本的规则组code
	compatMap := map[string]string{
		"k8s_cluster":      "k8s_container",    // 新code -> 旧code
		"process_metrics":  "process_resources", // 新code -> 旧code
		"other_metrics":    "others",            // 新code -> 旧code
	}

	// 定义默认规则
	defaultRules := []struct {
		GroupCode     string
		Name          string
		Type          bool
		ShowInTable   bool
		Description   string
		Query         string
		Threshold     *float64
		ThresholdType string
		Unit          string
		Labels        string
	}{
		// 服务器基础资源
		{"basic_resources", "CPU使用率", true, true, "节点CPU使用率统计", "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode='idle'}[5m])) * 100)", floatPtr(80), "greater", "%", `{"instance":"节点"}`},
		{"basic_resources", "CPU核心数", false, true, "节点CPU核心数统计", "count by (instance) (node_cpu_seconds_total{mode='idle'})", nil, "", "core", `{"instance":"节点"}`},
		{"basic_resources", "内存总量", false, true, "节点内存总量统计", "node_memory_MemTotal_bytes", nil, "", "B", `{"instance":"节点"}`},
		{"basic_resources", "内存使用量", false, true, "节点内存使用量统计", "node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes", nil, "", "B", `{"instance":"节点"}`},
		{"basic_resources", "内存使用率", true, true, "节点内存使用率统计", "100 - ((node_memory_MemAvailable_bytes * 100) / node_memory_MemTotal_bytes)", floatPtr(85), "greater", "%", `{"instance":"节点"}`},
		{"basic_resources", "磁盘总量", false, true, "节点磁盘总量统计", `node_filesystem_size_bytes{fstype=~"ext.*|xfs",mountpoint !~".*pod.*|/run.*|/boot.*|/tmp.*"}`, nil, "", "B", `{"instance":"节点","mountpoint":"挂载点","device":"磁盘"}`},
		{"basic_resources", "磁盘使用量", false, true, "节点磁盘使用量统计", `node_filesystem_size_bytes{fstype=~"ext.*|xfs",mountpoint !~".*pod.*|/run.*|/boot.*|/tmp.*"} - node_filesystem_avail_bytes{fstype=~"ext.*|xfs",mountpoint !~".*pod.*|/run.*|/boot.*|/tmp.*"}`, nil, "", "B", `{"instance":"节点","mountpoint":"挂载点","device":"磁盘"}`},
		{"basic_resources", "磁盘使用率", true, true, "节点磁盘使用率统计", `(node_filesystem_size_bytes{fstype=~"ext.*|xfs",mountpoint !~".*pod.*|/run.*|/boot.*|/tmp.*"}-node_filesystem_free_bytes{fstype=~"ext.*|xfs",mountpoint !~".*pod.*|/run.*|/boot.*|/tmp.*"}) *100/(node_filesystem_avail_bytes{fstype=~"ext.*|xfs",mountpoint !~".*pod.*|/run.*|/boot.*|/tmp.*"}+(node_filesystem_size_bytes{fstype=~"ext.*|xfs",mountpoint !~".*pod.*|/run.*|/boot.*|/tmp.*"}-node_filesystem_free_bytes{fstype=~"ext.*|xfs",mountpoint !~".*pod.*|/run.*|/boot.*|/tmp.*"}))`, floatPtr(80), "greater", "%", `{"instance":"节点","mountpoint":"挂载点","device":"磁盘"}`},
		{"basic_resources", "运行时间", false, true, "系统运行时长统计", "time() - node_boot_time_seconds", nil, "", "s", `{"instance":"节点"}`},
		{"basic_resources", "5分钟负载", false, true, "系统5分钟平均负载", "node_load5", nil, "", "", `{"instance":"节点"}`},
		{"basic_resources", "30分钟内磁盘平均读取值", true, false, "30分钟内磁盘平均读取速率", `avg_over_time(rate(node_disk_read_bytes_total{device=~"vd.*|sd.*"}[5m])[30m:1m]) / 1024 / 1024`, floatPtr(100), "greater", "MB/s", `{"instance":"节点","device":"设备"}`},
		{"basic_resources", "30分钟内磁盘平均写入值", true, false, "30分钟内磁盘平均写入速率", `avg_over_time(rate(node_disk_written_bytes_total{device=~"vd.*|sd.*"}[5m])[30m:1m]) / 1024 / 1024`, floatPtr(100), "greater", "MB/s", `{"instance":"节点","device":"设备"}`},
		{"basic_resources", "TCP连接数", false, true, "当前活跃的TCP连接总数", "node_netstat_Tcp_CurrEstab", nil, "", "个", `{"instance":"节点"}`},
		{"basic_resources", "TCP_TW数", false, true, "TCP TIME_WAIT状态连接数", "node_sockstat_TCP_tw", nil, "", "个", `{"instance":"节点"}`},
		{"basic_resources", "30分钟内下载速率", true, false, "30分钟内网络平均下载速率", `avg_over_time(rate(node_network_receive_bytes_total{device=~"eth.*|ens.*"}[5m])[30m:1m]) / 1024 / 1024`, floatPtr(100), "greater", "MB/s", `{"instance":"节点","device":"设备"}`},
		{"basic_resources", "30分钟内上传速率", true, false, "30分钟内网络平均上传速率", `avg_over_time(rate(node_network_transmit_bytes_total{device=~"eth.*|ens.*"}[5m])[30m:1m]) / 1024 / 1024`, floatPtr(100), "greater", "MB/s", `{"instance":"节点","device":"设备"}`},

		// Kubernetes集群状态
		{"k8s_cluster", "节点就绪状态", true, false, "K8s节点就绪状态检查", `kube_node_status_condition{condition='Ready',status!='true'}`, floatPtr(0), "equal", "", `{"node":"节点","condition":"状态类型"}`},
		{"k8s_cluster", "Kubelet证书状态", true, false, "kubelet证书有效期检查", "kubelet_cert_days_left", floatPtr(30), "at_least", "天", `{"instance":"节点"}`},
		{"k8s_cluster", "Kubeproxy证书状态", true, false, "kubeproxy证书有效期检查", "kubeproxy_cert_days_left", floatPtr(30), "at_least", "天", `{"instance":"节点"}`},
		{"k8s_cluster", "Kubecontroller证书状态", true, false, "kubecontroller证书有效期检查", "kube_controller_cert_days_left", floatPtr(30), "at_least", "天", `{"instance":"节点"}`},
		{"k8s_cluster", "Pod运行状态", true, false, "集群Pod运行状态统计", `sum by (namespace, pod) (kube_pod_status_phase{phase="Running"} == 1 and on(namespace, pod) kube_pod_owner{owner_kind=~"Deployment|ReplicaSet|StatefulSet|DaemonSet"})`, floatPtr(1), "equal", "", `{"namespace":"命名空间","pod":"Pod名称"}`},
		{"k8s_cluster", "PVC使用率", true, false, "持久化存储使用率", `100 * (1 - kubelet_volume_stats_available_bytes / kubelet_volume_stats_capacity_bytes)`, floatPtr(90), "greater", "%", `{"namespace":"命名空间","persistentvolumeclaim":"PVC名称"}`},

		// 进程指标
		{"process_metrics", "进程CPU使用率top5", false, false, "进程CPU使用率top5", "topk by (instance) (5, rate(namedprocess_namegroup_cpu_seconds_total[5m]))", nil, "", "%", `{"instance":"节点","groupname":"进程名"}`},
		{"process_metrics", "进程内存使用率top5", false, false, "进程内存使用率top5", `topk by (instance) (5,(avg_over_time(namedprocess_namegroup_memory_bytes{memtype="swapped"}[5m])+ ignoring (memtype) avg_over_time(namedprocess_namegroup_memory_bytes{memtype="resident"}[5m])) / (1024 * 1204))`, nil, "", "MB", `{"instance":"节点","groupname":"进程名"}`},

		// 其他指标
		{"other_metrics", "域名证书有效期小于30天", true, false, "域名证书有效期检查", "round((probe_ssl_earliest_cert_expiry - time()) / 86400)", floatPtr(60), "at_least", "天", `{"instance":"节点","target":"域名"}`},
	}

	// 创建规则
	rules := make([]Rule, 0, len(defaultRules))
	sortOrder := 0
	for _, r := range defaultRules {
		// 先尝试精确匹配
		groupID, ok := groupMap[r.GroupCode]
		if !ok {
			// 尝试使用兼容映射
			altCode, hasAlt := compatMap[r.GroupCode]
			if hasAlt {
				groupID, ok = groupMap[altCode]
			}
		}
		if !ok {
			continue
		}
		rules = append(rules, Rule{
			GroupID:       groupID,
			Name:          r.Name,
			Type:          r.Type,
			ShowInTable:   r.ShowInTable,
			Description:   r.Description,
			Query:         r.Query,
			Threshold:     r.Threshold,
			ThresholdType: r.ThresholdType,
			Unit:          r.Unit,
			Labels:        r.Labels,
			ProjectScope:  "*",
			Enabled:       true,
			SortOrder:     sortOrder,
		})
		sortOrder++
	}

	return DB.Create(&rules).Error
}

// floatPtr 辅助函数，创建 float64 指针
func floatPtr(v float64) *float64 {
	return &v
}
