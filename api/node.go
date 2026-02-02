package api

import (
	"bytes"
	"html/template"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node"
	"sublink/node/protocol"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func NodeUpdadte(c *gin.Context) {
	var Node models.Node
	name := c.PostForm("name")
	oldname := c.PostForm("oldname")
	oldlink := c.PostForm("oldlink")
	link := c.PostForm("link")
	dialerProxyName := c.PostForm("dialerProxyName")
	group := c.PostForm("group")
	if name == "" || link == "" {
		utils.FailWithMsg(c, "节点名称 or 备注不能为空")
		return
	}
	// 查找旧节点
	Node.Name = oldname
	Node.Link = oldlink
	err := Node.Find()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	Node.Name = name

	//更新构造节点元数据
	// 检测是否为 WireGuard 配置文件格式，如果是则转换为 URL 格式
	if protocol.IsWireGuardConfig(link) {
		wg, err := protocol.ParseWireGuardConfig(link)
		if err != nil {
			utils.FailWithMsg(c, "WireGuard 配置文件解析失败: "+err.Error())
			return
		}
		// 转换为 URL 格式
		link = protocol.EncodeWireGuardURL(wg)
	}
	u, err := url.Parse(link)
	if err != nil {
		utils.Error("解析节点链接失败: %v", err)
		return
	}
	switch {
	case u.Scheme == "ss":
		ss, err := protocol.DecodeSSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = ss.Name
		}
		Node.LinkName = ss.Name
		Node.LinkAddress = ss.Server + ":" + utils.GetPortString(ss.Port)
		Node.LinkHost = ss.Server
		Node.LinkPort = utils.GetPortString(ss.Port)
	case u.Scheme == "ssr":
		ssr, err := protocol.DecodeSSRURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = ssr.Qurey.Remarks
		}
		Node.LinkName = ssr.Qurey.Remarks
		Node.LinkAddress = ssr.Server + ":" + utils.GetPortString(ssr.Port)
		Node.LinkHost = ssr.Server
		Node.LinkPort = utils.GetPortString(ssr.Port)
	case u.Scheme == "trojan":
		trojan, err := protocol.DecodeTrojanURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}

		if Node.Name == "" {
			Node.Name = trojan.Name
		}
		Node.LinkName = trojan.Name
		Node.LinkAddress = trojan.Hostname + ":" + utils.GetPortString(trojan.Port)
		Node.LinkHost = trojan.Hostname
		Node.LinkPort = utils.GetPortString(trojan.Port)
	case u.Scheme == "vmess":
		vmess, err := protocol.DecodeVMESSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = vmess.Ps
		}
		Node.LinkName = vmess.Ps
		prot := utils.GetPortString(vmess.Port)
		Node.LinkAddress = vmess.Add + ":" + prot
		Node.LinkHost = vmess.Host
		Node.LinkPort = prot
	case u.Scheme == "vless":
		vless, err := protocol.DecodeVLESSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = vless.Name
		}
		Node.LinkName = vless.Name
		Node.LinkAddress = vless.Server + ":" + utils.GetPortString(vless.Port)
		Node.LinkHost = vless.Server
		Node.LinkPort = utils.GetPortString(vless.Port)
	case u.Scheme == "hy" || u.Scheme == "hysteria":
		hy, err := protocol.DecodeHYURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = hy.Name
		}
		Node.LinkName = hy.Name
		Node.LinkAddress = hy.Host + ":" + utils.GetPortString(hy.Port)
		Node.LinkHost = hy.Host
		Node.LinkPort = utils.GetPortString(hy.Port)
	case u.Scheme == "hy2" || u.Scheme == "hysteria2":
		hy2, err := protocol.DecodeHY2URL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = hy2.Name
		}
		Node.LinkName = hy2.Name
		Node.LinkAddress = hy2.Host + ":" + utils.GetPortString(hy2.Port)
		Node.LinkHost = hy2.Host
		Node.LinkPort = utils.GetPortString(hy2.Port)
	case u.Scheme == "tuic":
		tuic, err := protocol.DecodeTuicURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = tuic.Name
		}
		Node.LinkName = tuic.Name
		Node.LinkAddress = tuic.Host + ":" + utils.GetPortString(tuic.Port)
		Node.LinkHost = tuic.Host
		Node.LinkPort = utils.GetPortString(tuic.Port)
	case u.Scheme == "socks5":
		socks5, err := protocol.DecodeSocks5URL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = socks5.Name
		}
		Node.LinkName = socks5.Name
		Node.LinkAddress = socks5.Server + ":" + utils.GetPortString(socks5.Port)
		Node.LinkHost = socks5.Server
		Node.LinkPort = utils.GetPortString(socks5.Port)
	case u.Scheme == "http" || u.Scheme == "https":
		httpProxy, err := protocol.DecodeHTTPURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = httpProxy.Name
		}
		Node.LinkName = httpProxy.Name
		Node.LinkAddress = httpProxy.Server + ":" + utils.GetPortString(httpProxy.Port)
		Node.LinkHost = httpProxy.Server
		Node.LinkPort = utils.GetPortString(httpProxy.Port)
	case u.Scheme == "wg" || u.Scheme == "wireguard":
		wg, err := protocol.DecodeWireGuardURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = wg.Name
		}
		Node.LinkName = wg.Name
		Node.LinkAddress = wg.Server + ":" + utils.GetPortString(wg.Port)
		Node.LinkHost = wg.Server
		Node.LinkPort = utils.GetPortString(wg.Port)
	}

	Node.Link = link
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	Node.Protocol = protocol.GetProtocolFromLink(link)

	// 重新计算 ContentHash
	proxy, proxyErr := protocol.LinkToProxy(protocol.Urls{Url: link}, protocol.OutputConfig{})
	if proxyErr == nil {
		contentHash := protocol.GenerateProxyContentHash(proxy)
		if contentHash != "" {
			Node.ContentHash = contentHash
			// 检查是否与其他节点重复（排除自身）
			if existingNode, exists := models.GetNodeByContentHash(contentHash); exists && existingNode.ID != Node.ID {
				// 构建详细的重复信息
				source := existingNode.Source
				if source == "" || source == "manual" {
					source = "手动添加"
				}
				group := existingNode.Group
				if group == "" {
					group = "未分组"
				}
				utils.FailWithMsg(c, "节点内容已存在，与以下节点重复：[来源: "+source+"] [分组: "+group+"] [名称: "+existingNode.Name+"]")
				return
			}
		}
	}

	err = Node.Update()
	if err != nil {
		utils.FailWithMsg(c, "更新失败")
		return
	}

	// 处理标签
	tags := c.PostForm("tags")
	if tags != "" {
		tagNames := strings.Split(tags, ",")
		// 过滤空字符串
		var validTagNames []string
		for _, t := range tagNames {
			t = strings.TrimSpace(t)
			if t != "" {
				validTagNames = append(validTagNames, t)
			}
		}
		_ = Node.SetTagNames(validTagNames)
	} else {
		// 如果 tags 参数为空，清除标签
		_ = Node.SetTagNames([]string{})
	}

	utils.OkWithMsg(c, "更新成功")
}

// 获取节点列表
func NodeGet(c *gin.Context) {
	var Node models.Node

	// 解析过滤参数
	filter := models.NodeFilter{
		Search:      c.Query("search"),
		Group:       c.Query("group"),
		Source:      c.Query("source"),
		Protocol:    c.Query("protocol"),
		SpeedStatus: c.Query("speedStatus"),
		DelayStatus: c.Query("delayStatus"),
		SortBy:      c.Query("sortBy"),
		SortOrder:   c.Query("sortOrder"),
	}

	// 安全解析数值参数
	if maxDelayStr := c.Query("maxDelay"); maxDelayStr != "" {
		if maxDelay, err := strconv.Atoi(maxDelayStr); err == nil && maxDelay > 0 {
			filter.MaxDelay = maxDelay
		}
	}

	if minSpeedStr := c.Query("minSpeed"); minSpeedStr != "" {
		if minSpeed, err := strconv.ParseFloat(minSpeedStr, 64); err == nil && minSpeed > 0 {
			filter.MinSpeed = minSpeed
		}
	}

	// 解析国家代码数组
	filter.Countries = c.QueryArray("countries[]")

	// 解析标签数组
	filter.Tags = c.QueryArray("tags[]")

	// 验证排序字段（白名单）
	if filter.SortBy != "" && filter.SortBy != "delay" && filter.SortBy != "speed" {
		filter.SortBy = "" // 无效排序字段，忽略
	}

	// 验证排序顺序
	if filter.SortOrder != "" && filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		filter.SortOrder = "asc" // 默认升序
	}

	// 解析分页参数
	page := 0
	pageSize := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 如果提供了分页参数，返回分页响应
	if page > 0 && pageSize > 0 {
		nodes, total, err := Node.ListWithFiltersPaginated(filter, page, pageSize)
		if err != nil {
			utils.FailWithMsg(c, "node list error")
			return
		}
		totalPages := 0
		if pageSize > 0 {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
		utils.OkDetailed(c, "node get", gin.H{
			"items":      nodes,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}

	// 不带分页参数，返回全部（向后兼容）
	nodes, err := Node.ListWithFilters(filter)
	if err != nil {
		utils.FailWithMsg(c, "node list error")
		return
	}
	utils.OkDetailed(c, "node get", nodes)
}

// NodeGetIDs 获取符合过滤条件的所有节点ID（用于全选操作）
func NodeGetIDs(c *gin.Context) {
	var Node models.Node

	// 解析过滤参数
	filter := models.NodeFilter{
		Search:      c.Query("search"),
		Group:       c.Query("group"),
		Source:      c.Query("source"),
		Protocol:    c.Query("protocol"),
		SpeedStatus: c.Query("speedStatus"),
		DelayStatus: c.Query("delayStatus"),
		SortBy:      c.Query("sortBy"),
		SortOrder:   c.Query("sortOrder"),
	}

	// 安全解析数值参数
	if maxDelayStr := c.Query("maxDelay"); maxDelayStr != "" {
		if maxDelay, err := strconv.Atoi(maxDelayStr); err == nil && maxDelay > 0 {
			filter.MaxDelay = maxDelay
		}
	}

	if minSpeedStr := c.Query("minSpeed"); minSpeedStr != "" {
		if minSpeed, err := strconv.ParseFloat(minSpeedStr, 64); err == nil && minSpeed > 0 {
			filter.MinSpeed = minSpeed
		}
	}

	// 解析国家代码数组
	filter.Countries = c.QueryArray("countries[]")

	// 解析标签数组
	filter.Tags = c.QueryArray("tags[]")

	ids, err := Node.GetFilteredNodeIDs(filter)
	if err != nil {
		utils.FailWithMsg(c, "get node ids error")
		return
	}
	utils.OkDetailed(c, "node ids get", ids)
}

// 添加节点
// NodeAdd 添加节点
func NodeAdd(c *gin.Context) {
	var Node models.Node
	link := c.PostForm("link")
	name := c.PostForm("name")
	dialerProxyName := c.PostForm("dialerProxyName")
	group := c.PostForm("group")
	if link == "" {
		utils.FailWithMsg(c, "link 不能为空")
		return
	}
	// 检测是否为 WireGuard 配置文件格式，如果是则转换为 URL 格式
	if protocol.IsWireGuardConfig(link) {
		wg, err := protocol.ParseWireGuardConfig(link)
		if err != nil {
			utils.FailWithMsg(c, "WireGuard 配置文件解析失败: "+err.Error())
			return
		}
		// 转换为 URL 格式
		link = protocol.EncodeWireGuardURL(wg)
	}

	// 检测是否为 Clash YAML 配置格式
	if strings.Contains(link, "proxies:") {
		var clashConfig node.ClashConfig
		if err := yaml.Unmarshal([]byte(link), &clashConfig); err == nil && len(clashConfig.Proxies) > 0 {
			// 成功解析为 Clash YAML 格式，处理每个代理节点
			var addedCount, failedCount int
			for _, proxy := range clashConfig.Proxies {
				proxyLink := node.GenerateProxyLink(proxy)
				if proxyLink == "" {
					failedCount++
					continue
				}
				// 创建节点并添加
				var n models.Node
				n.Name = proxy.Name
				n.Link = proxyLink
				n.LinkName = proxy.Name
				n.LinkHost = proxy.Server
				n.LinkPort = strconv.Itoa(proxy.Port.Int())
				n.LinkAddress = proxy.Server + ":" + n.LinkPort
				n.DialerProxyName = dialerProxyName
				n.Group = group
				n.Protocol = proxy.Type

				// 生成 ContentHash
				contentHash := protocol.GenerateProxyContentHash(proxy)
				if contentHash != "" {
					n.ContentHash = contentHash
					// 检查是否已存在相同内容的节点
					if _, exists := models.GetNodeByContentHash(contentHash); exists {
						failedCount++
						continue
					}
				}

				if err := n.Add(); err != nil {
					failedCount++
					continue
				}
				addedCount++
			}

			if addedCount == 0 {
				utils.FailWithMsg(c, "Clash YAML 解析成功但无法添加任何节点（可能全部重复或格式不支持）")
				return
			}
			utils.OkWithMsg(c, "Clash YAML 导入完成，成功添加 "+strconv.Itoa(addedCount)+" 个节点")
			return
		}
	}

	if !strings.Contains(link, "://") {
		utils.FailWithMsg(c, "link 必须包含 :// 或者是有效的 WireGuard/Clash YAML 配置文件")
		return
	}

	// ============================================
	// 修改核心：调用公共解析函数
	// ============================================
	parsedNode, err := ParseNodeFromLink(link)
	if err != nil {
		utils.Error("解析节点链接失败: %v", err)
		utils.FailWithMsg(c, "解析节点链接失败")
		return
	}
	// 将解析结果赋值给当前的 Node 对象
	Node = parsedNode

	// 如果用户在表单手动输入了名称，覆盖解析出来的名称
	if name != "" {
		Node.Name = name
	}

	// 设置其他参数
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	// Node.Link 和 Protocol 已经在 ParseNodeFromLink 中赋值了

	// 生成 ContentHash（用于全库去重）
	proxy, proxyErr := protocol.LinkToProxy(protocol.Urls{Url: link}, protocol.OutputConfig{})
	if proxyErr == nil {
		contentHash := protocol.GenerateProxyContentHash(proxy)
		if contentHash != "" {
			Node.ContentHash = contentHash
			// 检查是否已存在相同内容的节点
			if existingNode, exists := models.GetNodeByContentHash(contentHash); exists {
				// 构建详细的重复信息
				source := existingNode.Source
				if source == "" || source == "manual" {
					source = "手动添加"
				}
				group := existingNode.Group
				if group == "" {
					group = "未分组"
				}
				utils.FailWithMsg(c, "节点内容已存在，与以下节点重复：[来源: "+source+"] [分组: "+group+"] [名称: "+existingNode.Name+"]")
				return
			}
		}
	}

	err = Node.Add()
	if err != nil {
		utils.FailWithMsg(c, "添加失败检查一下是否节点重复")
		return
	}

	// 处理标签
	tags := c.PostForm("tags")
	if tags != "" {
		tagNames := strings.Split(tags, ",")
		// 过滤空字符串
		var validTagNames []string
		for _, t := range tagNames {
			t = strings.TrimSpace(t)
			if t != "" {
				validTagNames = append(validTagNames, t)
			}
		}
		_ = Node.SetTagNames(validTagNames)
	}

	utils.OkWithMsg(c, "添加成功")
}

// 删除节点
func NodeDel(c *gin.Context) {
	var Node models.Node
	id := c.Query("id")
	if id == "" {
		utils.FailWithMsg(c, "id 不能为空")
		return
	}
	x, _ := strconv.Atoi(id)
	Node.ID = x
	err := Node.Del()
	if err != nil {
		utils.FailWithMsg(c, "删除失败")
		return
	}
	utils.OkWithMsg(c, "删除成功")
}

// 节点统计
func NodesTotal(c *gin.Context) {
	var Node models.Node
	nodes, err := Node.List()
	if err != nil {
		utils.FailWithMsg(c, "获取不到节点统计")
		return
	}

	total := len(nodes)
	available := 0
	for _, n := range nodes {
		if n.Speed > 0 && n.DelayTime > 0 {
			available++
		}
	}

	utils.OkDetailed(c, "取得节点统计", gin.H{
		"total":     total,
		"available": available,
	})
}

// NodeBatchDel 批量删除节点
func NodeBatchDel(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要删除的节点")
		return
	}
	err := models.BatchDel(req.IDs)
	if err != nil {
		utils.FailWithMsg(c, "批量删除失败")
		return
	}
	utils.OkWithMsg(c, "批量删除成功")
}

// NodeBatchUpdateGroup 批量更新节点分组
func NodeBatchUpdateGroup(c *gin.Context) {
	var req struct {
		IDs   []int  `json:"ids"`
		Group string `json:"group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要修改的节点")
		return
	}
	err := models.BatchUpdateGroup(req.IDs, req.Group)
	if err != nil {
		utils.FailWithMsg(c, "批量更新分组失败")
		return
	}
	utils.OkWithMsg(c, "批量更新分组成功")
}

// NodeBatchUpdateDialerProxy 批量更新节点前置代理
func NodeBatchUpdateDialerProxy(c *gin.Context) {
	var req struct {
		IDs             []int  `json:"ids"`
		DialerProxyName string `json:"dialerProxyName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要修改的节点")
		return
	}
	err := models.BatchUpdateDialerProxy(req.IDs, req.DialerProxyName)
	if err != nil {
		utils.FailWithMsg(c, "批量更新前置代理失败")
		return
	}
	utils.OkWithMsg(c, "批量更新前置代理成功")
}

// NodeBatchUpdateSource 批量更新节点来源
func NodeBatchUpdateSource(c *gin.Context) {
	var req struct {
		IDs    []int  `json:"ids"`
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要修改的节点")
		return
	}
	err := models.BatchUpdateSource(req.IDs, req.Source)
	if err != nil {
		utils.FailWithMsg(c, "批量更新来源失败")
		return
	}
	utils.OkWithMsg(c, "批量更新来源成功")
}

// NodeBatchUpdateCountry 批量更新节点国家代码
func NodeBatchUpdateCountry(c *gin.Context) {
	var req struct {
		IDs     []int  `json:"ids"`
		Country string `json:"country"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要修改的节点")
		return
	}
	// 国家代码转大写，保持一致性
	country := strings.ToUpper(strings.TrimSpace(req.Country))
	err := models.BatchUpdateCountry(req.IDs, country)
	if err != nil {
		utils.FailWithMsg(c, "批量更新国家代码失败")
		return
	}
	utils.OkWithMsg(c, "批量更新国家代码成功")
}

// 获取所有分组列表
func GetGroups(c *gin.Context) {
	var node models.Node
	groups, err := node.GetAllGroups()
	if err != nil {
		utils.FailWithMsg(c, "获取分组列表失败")
		return
	}
	utils.OkDetailed(c, "获取分组列表成功", groups)
}

// GetSources 获取所有来源列表
func GetSources(c *gin.Context) {
	var node models.Node
	sources, err := node.GetAllSources()
	if err != nil {
		utils.FailWithMsg(c, "获取来源列表失败")
		return
	}
	utils.OkDetailed(c, "获取来源列表成功", sources)
}

// FastestSpeedNode 获取最快速度节点
func FastestSpeedNode(c *gin.Context) {
	node := models.GetFastestSpeedNode()
	utils.OkDetailed(c, "获取最快速度节点成功", node)
}

// LowestDelayNode 获取最低延迟节点
func LowestDelayNode(c *gin.Context) {
	node := models.GetLowestDelayNode()
	utils.OkDetailed(c, "获取最低延迟节点成功", node)
}

// GetNodeCountries 获取所有节点的国家代码列表
func GetNodeCountries(c *gin.Context) {
	countries := models.GetAllCountries()
	utils.OkDetailed(c, "获取国家代码成功", countries)
}

// NodeCountryStats 获取按国家统计的节点数量
func NodeCountryStats(c *gin.Context) {
	stats := models.GetNodeCountryStats()
	utils.OkDetailed(c, "获取国家统计成功", stats)
}

// NodeProtocolStats 获取按协议统计的节点数量
func NodeProtocolStats(c *gin.Context) {
	stats := models.GetNodeProtocolStats()
	utils.OkDetailed(c, "获取协议统计成功", stats)
}

// NodeTagStats 获取按标签统计的节点数量
func NodeTagStats(c *gin.Context) {
	stats := models.GetNodeTagStats()
	utils.OkDetailed(c, "获取标签统计成功", stats)
}

// NodeGroupStats 获取按分组统计的节点数量
func NodeGroupStats(c *gin.Context) {
	stats := models.GetNodeGroupStats()
	utils.OkDetailed(c, "获取分组统计成功", stats)
}

// NodeSourceStats 获取按来源统计的节点数量
func NodeSourceStats(c *gin.Context) {
	stats := models.GetNodeSourceStats()
	utils.OkDetailed(c, "获取来源统计成功", stats)
}

// GetIPDetails 获取IP详细信息
// GET /api/v1/nodes/ip-info?ip=xxx.xxx.xxx.xxx
func GetIPDetails(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		utils.FailWithMsg(c, "IP地址不能为空")
		return
	}

	// 调用模型层获取IP信息（多级缓存）
	ipInfo, err := models.GetIPInfo(ip)
	if err != nil {
		utils.FailWithMsg(c, "查询IP信息失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "获取成功", ipInfo)
}

// GetIPCacheStats 获取IP缓存统计
// GET /api/v1/nodes/ip-cache/stats
func GetIPCacheStats(c *gin.Context) {
	count := models.GetIPInfoCount()
	utils.OkDetailed(c, "获取成功", gin.H{
		"count": count,
	})
}

// ClearIPCache 清除所有IP缓存
// DELETE /api/v1/nodes/ip-cache
func ClearIPCache(c *gin.Context) {
	err := models.ClearAllIPInfo()
	if err != nil {
		utils.FailWithMsg(c, "清除失败: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "IP缓存已清除")
}

// GetNodeProtocols 获取所有使用中的协议类型列表（用于过滤器选项）
// GET /api/v1/nodes/protocols
func GetNodeProtocols(c *gin.Context) {
	protocols := models.GetAllProtocols()
	utils.OkDetailed(c, "获取协议列表成功", protocols)
}

// ParseNodeFromLink [公共函数] 将链接解析为 Node 结构体
// 提取自 NodeAdd，供 AutoReport 和 NodeAdd 复用
func ParseNodeFromLink(link string) (models.Node, error) {
	var Node models.Node

	u, err := url.Parse(link)
	if err != nil {
		return Node, err
	}

	switch {
	case u.Scheme == "ss":
		ss, err := protocol.DecodeSSURL(link)
		if err == nil {
			Node.Name = ss.Name
			Node.LinkName = ss.Name
			Node.LinkAddress = ss.Server + ":" + utils.GetPortString(ss.Port)
			Node.LinkHost = ss.Server
			Node.LinkPort = utils.GetPortString(ss.Port)
		}
	case u.Scheme == "ssr":
		ssr, err := protocol.DecodeSSRURL(link)
		if err == nil {
			Node.Name = ssr.Qurey.Remarks
			Node.LinkName = ssr.Qurey.Remarks
			Node.LinkAddress = ssr.Server + ":" + utils.GetPortString(ssr.Port)
			Node.LinkHost = ssr.Server
			Node.LinkPort = utils.GetPortString(ssr.Port)
		}
	case u.Scheme == "trojan":
		trojan, err := protocol.DecodeTrojanURL(link)
		if err == nil {
			Node.Name = trojan.Name
			Node.LinkName = trojan.Name
			Node.LinkAddress = trojan.Hostname + ":" + utils.GetPortString(trojan.Port)
			Node.LinkHost = trojan.Hostname
			Node.LinkPort = utils.GetPortString(trojan.Port)
		}
	case u.Scheme == "vmess":
		vmess, err := protocol.DecodeVMESSURL(link)
		if err == nil {
			Node.Name = vmess.Ps
			Node.LinkName = vmess.Ps
			port := utils.GetPortString(vmess.Port)
			Node.LinkAddress = vmess.Add + ":" + port
			Node.LinkHost = vmess.Host
			Node.LinkPort = port
		}
	case u.Scheme == "vless":
		vless, err := protocol.DecodeVLESSURL(link)
		if err == nil {
			Node.Name = vless.Name
			Node.LinkName = vless.Name
			Node.LinkAddress = vless.Server + ":" + utils.GetPortString(vless.Port)
			Node.LinkHost = vless.Server
			Node.LinkPort = utils.GetPortString(vless.Port)
		}
	case u.Scheme == "hy" || u.Scheme == "hysteria":
		hy, err := protocol.DecodeHYURL(link)
		if err == nil {
			Node.Name = hy.Name
			Node.LinkName = hy.Name
			Node.LinkAddress = hy.Host + ":" + utils.GetPortString(hy.Port)
			Node.LinkHost = hy.Host
			Node.LinkPort = utils.GetPortString(hy.Port)
		}
	case u.Scheme == "hy2" || u.Scheme == "hysteria2":
		hy2, err := protocol.DecodeHY2URL(link)
		if err == nil {
			Node.Name = hy2.Name
			Node.LinkName = hy2.Name
			Node.LinkAddress = hy2.Host + ":" + utils.GetPortString(hy2.Port)
			Node.LinkHost = hy2.Host
			Node.LinkPort = utils.GetPortString(hy2.Port)
		}
	case u.Scheme == "tuic":
		tuic, err := protocol.DecodeTuicURL(link)
		if err == nil {
			Node.Name = tuic.Name
			Node.LinkName = tuic.Name
			Node.LinkAddress = tuic.Host + ":" + utils.GetPortString(tuic.Port)
			Node.LinkHost = tuic.Host
			Node.LinkPort = utils.GetPortString(tuic.Port)
		}
	case u.Scheme == "socks5":
		socks5, err := protocol.DecodeSocks5URL(link)
		if err == nil {
			Node.Name = socks5.Name
			Node.LinkName = socks5.Name
			Node.LinkAddress = socks5.Server + ":" + utils.GetPortString(socks5.Port)
			Node.LinkHost = socks5.Server
			Node.LinkPort = utils.GetPortString(socks5.Port)
		}
	case u.Scheme == "anytls":
		anytls, err := protocol.DecodeAnyTLSURL(link)
		if err == nil {
			Node.Name = anytls.Name
			Node.LinkName = anytls.Name
			Node.LinkAddress = anytls.Server + ":" + utils.GetPortString(anytls.Port)
			Node.LinkHost = anytls.Server
			Node.LinkPort = utils.GetPortString(anytls.Port)
		}
	case u.Scheme == "wg" || u.Scheme == "wireguard":
		wg, err := protocol.DecodeWireGuardURL(link)
		if err == nil {
			Node.Name = wg.Name
			Node.LinkName = wg.Name
			Node.LinkAddress = wg.Server + ":" + utils.GetPortString(wg.Port)
			Node.LinkHost = wg.Server
			Node.LinkPort = utils.GetPortString(wg.Port)
		}
	case u.Scheme == "http" || u.Scheme == "https":
		httpProxy, err := protocol.DecodeHTTPURL(link)
		if err == nil {
			Node.Name = httpProxy.Name
			Node.LinkName = httpProxy.Name
			Node.LinkAddress = httpProxy.Server + ":" + utils.GetPortString(httpProxy.Port)
			Node.LinkHost = httpProxy.Server
			Node.LinkPort = utils.GetPortString(httpProxy.Port)
		}
	}

	Node.Link = link
	Node.Protocol = protocol.GetProtocolFromLink(link)

	return Node, nil
}

// NodeReport 节点自动上报接口 (修改版：增加 Token 验证)
func NodeReport(c *gin.Context) {
	// 1. 定义请求结构体，增加 Token 字段
	var req struct {
		Link  string `json:"link" form:"link"`
		Group string `json:"group" form:"group"`
		Token string `json:"token" form:"token"` // 新增 Token 字段
	}

	// 2. 绑定参数
	if err := c.ShouldBind(&req); err != nil {
		req.Link = c.PostForm("link")
		req.Group = c.PostForm("group")
		req.Token = c.PostForm("token")
	}

	// === 核心修改：Token 验证 ===
	// 从数据库获取设置的 Token (假设 Settings 表中有 report_token 字段)
	// 如果你没有 models.GetSetting，请根据你的代码替换为实际获取系统配置的方法
	storedToken, _ := models.GetSetting("report_token")

	// 如果系统设置了 Token，则强制验证
	if storedToken != "" {
		if req.Token != storedToken {
			utils.FailWithMsg(c, "Invalid Report Token")
			return
		}
	}
	// ===========================

	// 设置默认分组
	if req.Group == "" {
		req.Group = "AutoReport"
	}

	if req.Link == "" {
		utils.FailWithMsg(c, "Link cannot be empty")
		return
	}

	// 3. 调用公共解析函数
	parsedNode, err := ParseNodeFromLink(req.Link)
	if err != nil {
		utils.FailWithMsg(c, "Invalid Link Format")
		return
	}

	// 4. 设置自动上报的特定参数
	parsedNode.Source = "自动上报"            // 标记来源
	parsedNode.Group = req.Group          // 使用获取到的分组
	parsedNode.DialerProxyName = "Direct" // 默认直连

	// 5. 生成 ContentHash (用于去重)
	if proxy, err := protocol.LinkToProxy(protocol.Urls{Url: req.Link}, protocol.OutputConfig{}); err == nil {
		parsedNode.ContentHash = protocol.GenerateProxyContentHash(proxy)
	}

	// 6. 核心：调用 UpsertNode (存在即更新，不存在即新增)
	if err := parsedNode.UpsertNode(); err != nil {
		utils.FailWithMsg(c, "Report Failed: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "Report Success")
}

// GetReportToken 获取节点上报 Token
func GetReportToken(c *gin.Context) {
	// 从数据库获取配置 (复用 models.GetSetting)
	token, _ := models.GetSetting("report_token")
	utils.OkWithData(c, gin.H{"reportToken": token})
}

// UpdateReportToken 更新节点上报 Token
func UpdateReportToken(c *gin.Context) {
	var req struct {
		ReportToken string `json:"reportToken"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	// 保存配置到数据库 (复用 models.SetSetting)
	err := models.SetSetting("report_token", req.ReportToken)
	if err != nil {
		utils.FailWithMsg(c, "保存失败")
		return
	}
	utils.OkWithMsg(c, "保存成功")
}

// getOrSetDefault 获取设置，如果不存在则设置默认值并保存到数据库
func getOrSetDefault(key string, defaultValue string) string {
	val, err := models.GetSetting(key)
	// 如果获取出错或值为空，则视为不存在（或需要初始化）
	if err != nil || val == "" {
		// 保存默认值到数据库
		// 注意：这里的 SetSetting 必须是你之前修复过的支持 Upsert 的版本
		if err := models.SetSetting(key, defaultValue); err != nil {
			utils.Error("初始化默认设置失败 [%s]: %v", key, err)
		}
		return defaultValue
	}
	return val
}

// GetInstallScript 动态生成安装脚本
// GET /subscription/install-singbox.sh
func GetInstallScript(c *gin.Context) {
	// 1. 准备默认配置数据
	// 如果数据库中没有这些 key，会自动写入默认值
	data := gin.H{
		"FixedPortSS":      getOrSetDefault("fixed_port_ss", "10001"),
		"FixedPortHY2":     getOrSetDefault("fixed_port_hy2", "10002"),
		"FixedPortTUIC":    getOrSetDefault("fixed_port_tuic", "10003"),
		"FixedPortReality": getOrSetDefault("fixed_port_reality", "10004"),
		"FixedPortSocks5":  getOrSetDefault("fixed_port_socks5", "10005"),

		"FixedRealitySNI": getOrSetDefault("fixed_reality_sni", "learn.microsoft.com"),
		"FixedSSMethod":   getOrSetDefault("fixed_ss_method", "2022-blake3-aes-128-gcm"),
		"FixedSocks5User": getOrSetDefault("fixed_socks5_user", "admin"),

		"FixedSocks5Pass": getOrSetDefault("fixed_socks5_pass", "nodeReporttest"),
	}

	// 2. 处理上报地址和 Token
	// 获取当前请求的 host (例如: example.com 或 1.2.3.4:8000)
	scheme := "http"
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := scheme + "://" + c.Request.Host

	// 如果数据库配置了 system_domain，优先使用它
	if domain, _ := models.GetSetting("system_domain"); domain != "" {
		if !strings.HasPrefix(domain, "http") {
			baseURL = "https://" + domain
		} else {
			baseURL = domain
		}
	}
	// 去除末尾斜杠
	baseURL = strings.TrimRight(baseURL, "/")

	// 设置默认上报地址
	data["ReportURL"] = baseURL + "/api/v1/nodes/report"

	// 获取 Token (如果没有设置则为空，脚本会留空)
	data["ReportToken"] = ""

	// 3. 读取模板文件
	// 注意：根据你的 Dockerfile，模板目录是 /app/template
	// 本地开发时可能是 ./template
	scriptPath := "template/install-singbox.sh"
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		utils.Error("读取脚本模板失败: %v", err)
		utils.FailWithMsg(c, "Script template not found")
		return
	}

	// 4. 解析并渲染模板
	tmpl, err := template.New("install-script").Parse(string(content))
	if err != nil {
		utils.Error("解析脚本模板失败: %v", err)
		utils.FailWithMsg(c, "Script template parse error")
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		utils.Error("渲染脚本模板失败: %v", err)
		utils.FailWithMsg(c, "Script render error")
		return
	}

	// 5. 返回结果 (Content-Type 设置为 shell 脚本)
	// 禁止缓存，确保每次获取都是最新的配置
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Data(200, "application/x-sh; charset=utf-8", buf.Bytes())
}

// GetInstallScriptConfig 获取安装脚本的配置参数
// GET /api/v1/nodes/install-config
func GetInstallScriptConfig(c *gin.Context) {
	// 复用 getOrSetDefault 确保数据库里有值（如果没有则初始化为默认值）
	config := gin.H{
		"fixedPortSS":      getOrSetDefault("fixed_port_ss", "10001"),
		"fixedPortHY2":     getOrSetDefault("fixed_port_hy2", "10002"),
		"fixedPortTUIC":    getOrSetDefault("fixed_port_tuic", "10003"),
		"fixedPortReality": getOrSetDefault("fixed_port_reality", "10004"),
		"fixedPortSocks5":  getOrSetDefault("fixed_port_socks5", "10005"),

		"fixedRealitySNI": getOrSetDefault("fixed_reality_sni", "learn.microsoft.com"),
		"fixedSSMethod":   getOrSetDefault("fixed_ss_method", "2022-blake3-aes-128-gcm"),
		"fixedSocks5User": getOrSetDefault("fixed_socks5_user", "admin"),
		"fixedSocks5Pass": getOrSetDefault("fixed_socks5_pass", "nodeReporttest"),
	}

	utils.OkDetailed(c, "获取配置成功", config)
}

// UpdateInstallScriptConfig 更新安装脚本的配置参数
// POST /api/v1/nodes/install-config
func UpdateInstallScriptConfig(c *gin.Context) {
	var req struct {
		FixedPortSS      string `json:"fixedPortSS"`
		FixedPortHY2     string `json:"fixedPortHY2"`
		FixedPortTUIC    string `json:"fixedPortTUIC"`
		FixedPortReality string `json:"fixedPortReality"`
		FixedPortSocks5  string `json:"fixedPortSocks5"`
		FixedRealitySNI  string `json:"fixedRealitySNI"`
		FixedSSMethod    string `json:"fixedSSMethod"`
		FixedSocks5User  string `json:"fixedSocks5User"`
		FixedSocks5Pass  string `json:"fixedSocks5Pass"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	// 辅助函数：保存配置并处理错误
	save := func(key, value string) error {
		// 允许为空值吗？如果业务逻辑允许空值，直接保存；如果不允许，可以在这里加判断
		return models.SetSetting(key, value)
	}

	// 逐个保存字段
	// 注意：前端传来的 key (如 fixedPortSS) 对应数据库的 key (如 fixed_port_ss)
	if err := save("fixed_port_ss", req.FixedPortSS); err != nil {
		utils.FailWithMsg(c, "保存 SS 端口失败")
		return
	}
	if err := save("fixed_port_hy2", req.FixedPortHY2); err != nil {
		utils.FailWithMsg(c, "保存 HY2 端口失败")
		return
	}
	if err := save("fixed_port_tuic", req.FixedPortTUIC); err != nil {
		utils.FailWithMsg(c, "保存 TUIC 端口失败")
		return
	}
	if err := save("fixed_port_reality", req.FixedPortReality); err != nil {
		utils.FailWithMsg(c, "保存 Reality 端口失败")
		return
	}
	if err := save("fixed_port_socks5", req.FixedPortSocks5); err != nil {
		utils.FailWithMsg(c, "保存 Socks5 端口失败")
		return
	}

	if err := save("fixed_reality_sni", req.FixedRealitySNI); err != nil {
		utils.FailWithMsg(c, "保存 SNI 失败")
		return
	}
	if err := save("fixed_ss_method", req.FixedSSMethod); err != nil {
		utils.FailWithMsg(c, "保存 SS 加密方式失败")
		return
	}
	if err := save("fixed_socks5_user", req.FixedSocks5User); err != nil {
		utils.FailWithMsg(c, "保存 Socks5 账号失败")
		return
	}
	if err := save("fixed_socks5_pass", req.FixedSocks5Pass); err != nil {
		utils.FailWithMsg(c, "保存 Socks5 密码失败")
		return
	}

	utils.OkWithMsg(c, "保存配置成功，新下载的脚本将应用此配置")
}
