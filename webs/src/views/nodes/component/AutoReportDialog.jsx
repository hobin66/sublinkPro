import { useState, useMemo, useEffect } from 'react';
import PropTypes from 'prop-types';

// material-ui
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import FormGroup from '@mui/material/FormGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import Checkbox from '@mui/material/Checkbox';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Tooltip from '@mui/material/Tooltip';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import InputAdornment from '@mui/material/InputAdornment';
import Alert from '@mui/material/Alert';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Chip from '@mui/material/Chip';
// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import CheckIcon from '@mui/icons-material/Check';
import TerminalIcon from '@mui/icons-material/Terminal';
import SettingsIcon from '@mui/icons-material/Settings';
import CodeIcon from '@mui/icons-material/Code';
import SaveIcon from '@mui/icons-material/Save';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import VpnKeyIcon from '@mui/icons-material/VpnKey';
import RemoveIcon from '@mui/icons-material/Remove';
// api
import { getSystemDomain } from 'api/settings';
import { 
  getReportToken, 
  updateReportToken, 
  getInstallScriptConfig, 
  updateInstallScriptConfig 
} from 'api/nodes';

const PROTOCOL_OPTIONS = [
  { label: 'VLESS', value: 'vless' },
  { label: 'Shadowsocks', value: 'shadowsocks' },
  { label: 'Hysteria2', value: 'hysteria2' },
  { label: 'Tuic', value: 'tuic' },
  { label: 'Socks5', value: 'socks5' }
];

// 自定义 TabPanel 组件
function TabPanel(props) {
  const { children, value, index, ...other } = props;
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && (
        // 🔧 调整此处 pt (padding-top) 数值可改变 Tab 标签栏与下方内容的间距
        // 当前: 2 (即 16px), 之前是 3 (24px)
        <Box sx={{ pt: 2, px: 3, pb: 3 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

TabPanel.propTypes = {
  children: PropTypes.node,
  index: PropTypes.number.isRequired,
  value: PropTypes.number.isRequired,
};

const AutoReportDialog = ({ open, onClose }) => {
  const [tabValue, setTabValue] = useState(0); 
  const [baseUrl, setBaseUrl] = useState(window.location.origin);
  const [token, setToken] = useState('');
  const [selectedProtocols, setSelectedProtocols] = useState(['vless', 'shadowsocks', 'hysteria2']);
  
  const [loading, setLoading] = useState(false);
  const [configLoading, setConfigLoading] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const [config, setConfig] = useState({
    fixedPortSS: '',
    fixedPortHY2: '',
    fixedPortTUIC: '',
    fixedPortReality: '',
    fixedPortSocks5: '',
    fixedRealitySNI: '',
    fixedSSMethod: '',
    fixedSocks5User: '',
    fixedSocks5Pass: ''
  });

  // 初始化加载
  useEffect(() => {
    if (open) {
      getSystemDomain().then((res) => {
        if (res.data && res.data.systemDomain) {
          let domain = res.data.systemDomain;
          if (!domain.startsWith('http')) domain = `https://${domain}`;
          setBaseUrl(domain);
        }
      });
      fetchToken();
      fetchConfig();
    }
  }, [open]);

  const fetchToken = async () => {
    try {
      const res = await getReportToken();
      setToken(res.data?.reportToken || '');
    } catch (err) {
      console.error("获取Token失败", err);
    }
  };

  const fetchConfig = async () => {
    setConfigLoading(true);
    try {
      const res = await getInstallScriptConfig();
      if (res.data) {
        setConfig(res.data);
      }
    } catch (err) {
      console.error("获取配置失败", err);
      showMsg('获取默认配置失败', 'error');
    } finally {
      setConfigLoading(false);
    }
  };

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  const handleCloseSnackbar = () => setSnackbar({ ...snackbar, open: false });
  const showMsg = (msg, severity = 'success') => setSnackbar({ open: true, message: msg, severity });

  const generatedCommand = useMemo(() => {
    const host = baseUrl.replace(/\/$/, '');
    const protocolStr = selectedProtocols.length > 0 ? selectedProtocols.join(' ') : 'vless';
    const tokenStr = token ? `--token "${token}"` : '';
    return `bash -c "$(curl -fsSL ${host}/report-add)" -- ${protocolStr} ${tokenStr}`;
  }, [baseUrl, token, selectedProtocols]);

  const handleCopy = async () => {
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(generatedCommand);
      } else {
        const textArea = document.createElement("textarea");
        textArea.value = generatedCommand;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
      }
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
      showMsg('复制成功', 'success');
    } catch (err) {
      showMsg('复制失败', 'error');
    }
  };

  const handleResetToken = async () => {
    setLoading(true);
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let newToken = '';
    for (let i = 0; i < 16; i++) newToken += chars.charAt(Math.floor(Math.random() * chars.length));

    try {
      await updateReportToken({ reportToken: newToken });
      setToken(newToken);
      showMsg('Token 已重置并保存', 'success');
    } catch (error) {
      showMsg('重置失败: ' + error.message, 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveConfig = async () => {
    setLoading(true);
    try {
      await updateInstallScriptConfig(config);
      showMsg('配置已保存', 'success');
    } catch (error) {
      showMsg('保存失败: ' + error.message, 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleConfigChange = (prop) => (event) => {
    setConfig({ ...config, [prop]: event.target.value });
  };

  const handleProtocolChange = (value) => {
    const currentIndex = selectedProtocols.indexOf(value);
    const newChecked = [...selectedProtocols];
    if (currentIndex === -1) newChecked.push(value);
    else newChecked.splice(currentIndex, 1);
    setSelectedProtocols(newChecked);
  };

const PROTOCOL_CONFIGS = [
    {
      id: 'ss',
      name: 'Shadowsocks',
      portField: 'fixedPortSS',
      hasExtra: true,
      transports: ['TCP', 'UDP'],
      extraFields: [
        { 
          label: '加密方式', 
          field: 'fixedSSMethod', 
          fullWidth: true,
          type: 'toggle',
          options: [
            { value: '2022-blake3-aes-128-gcm', label: '2022-blake3-aes-128-gcm' },
            { value: 'aes-128-gcm', label: 'AES-128-GCM' }
          ]
        }
      ]
    },
    {
      id: 'reality',
      name: 'VLESS Reality',
      portField: 'fixedPortReality',
      hasExtra: true,
      transports: ['TCP'],
      extraFields: [{ label: '伪装域名 (SNI)', field: 'fixedRealitySNI', fullWidth: true }]
    },
    { 
      id: 'hy2', 
      name: 'Hysteria2', 
      portField: 'fixedPortHY2', 
      hasExtra: false,
      transports: ['UDP']
    },
    { 
      id: 'tuic', 
      name: 'TUIC', 
      portField: 'fixedPortTUIC', 
      hasExtra: false,
      transports: ['UDP']
    },
    {
      id: 'socks5',
      name: 'Socks5',
      portField: 'fixedPortSocks5',
      hasExtra: true,
      transports: ['TCP', 'UDP'],
      extraFields: [{ label: '默认账号', field: 'fixedSocks5User' }, { label: '默认密码', field: 'fixedSocks5Pass' }]
    }
  ];

  return (
    <>
      <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
        <DialogTitle 
          sx={{ 
            m: 0, 
            p: 3, // 建议稍微增加一点内边距，因为去掉了图标，太窄可能不好看
            // display: 'flex', // 如果只有文字，flex 布局不再强制需要，但保留也无妨
            // alignItems: 'center', 
          }}
        >
          {/* [修改点 1] 已删除 TerminalIcon 图标 */}
          
          {/* [修改点 2] 调整 variant 或 sx 来改变字体大小 */}
          {/* variant="h4" 会比原来的 h6 大很多；也可以用 sx={{ fontSize: '1.5rem', fontWeight: 'bold' }} 自定义 */}
          <Typography variant="h4" component="div" sx={{ fontWeight: 'bold',fontSize: '1.15rem', }}>
            自动上报
          </Typography>
        </DialogTitle>
        
        <DialogContent dividers sx={{ p: 0 }}>
          <Box sx={{ width: '100%' }}>
            <Box sx={{ borderBottom: 1, borderColor: 'divider', px: 2 }}>
              <Tabs value={tabValue} onChange={handleTabChange} variant="fullWidth">
                <Tab icon={<CodeIcon />} iconPosition="start" label="安装命令" />
                <Tab icon={<SettingsIcon />} iconPosition="start" label="参数配置" />
              </Tabs>
            </Box>

            {/* === Tab 1: 安装命令 === */}
            <TabPanel value={tabValue} index={0}>
              <Stack spacing={3}> {/* 🔧 调整 spacing 可改变内部元素间距 */}
                
                {/* 协议选择 */}
                <Box>
                  <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                    选择协议:
                  </Typography>
                  <FormGroup row sx={{ justifyContent: 'space-between' }}>
                    {PROTOCOL_OPTIONS.map((option) => (
                      <FormControlLabel
                        key={option.value}
                        control={
                          <Checkbox
                            size="small"
                            checked={selectedProtocols.indexOf(option.value) !== -1}
                            onChange={() => handleProtocolChange(option.value)}
                          />
                        }
                        label={<Typography variant="body2">{option.label}</Typography>}
                      />
                    ))}
                  </FormGroup>
                </Box>

                {/* 命令代码块 */}
                <Box>
                  <Typography variant="subtitle2" color="primary" gutterBottom>一键安装命令 (点击复制)</Typography>
                  
                  <Paper
                    elevation={0}
                    sx={{
                      position: 'relative',
                      bgcolor: '#f5f5f5',
                      color: '#333',
                      p: 2,
                      borderRadius: 2,
                      fontFamily: '"JetBrains Mono", monospace',
                      fontSize: '0.85rem',
                      lineHeight: 1.6,
                      wordBreak: 'break-all',
                      cursor: 'pointer',
                      border: '1px solid #e0e0e0',
                      // 核心悬停效果：悬停时边框变色，且让内部的 .copy-btn 显示
                      '&:hover': { 
                        borderColor: '#bdbdbd',
                        '& .copy-btn': { opacity: 1 } // 🔧 鼠标移入时显示图标
                      }
                    }}
                    onClick={handleCopy}
                  >
                    <Tooltip title={copySuccess ? "已复制!" : "点击复制"}>
                      <IconButton
                        className="copy-btn" // 标记类名
                        size="small"
                        sx={{ 
                          position: 'absolute', 
                          top: 8, 
                          right: 8,
                          opacity: 0, // 🔧 默认透明度为 0 (隐藏)
                          transition: 'opacity 0.2s', // 平滑过渡
                          bgcolor: 'rgba(255,255,255,0.8)',
                          '&:hover': { bgcolor: '#fff' }
                        }}
                      >
                        {copySuccess ? <CheckIcon fontSize="small" color="success" /> : <ContentCopyIcon fontSize="small" />}
                      </IconButton>
                    </Tooltip>

                    <span style={{ color: '#005cc5' }}>bash</span> -c <span style={{ color: '#22863a' }}>"$(curl -fsSL {baseUrl}/report-add)"</span> -- 
                    <span style={{ color: '#6f42c1' }}> {selectedProtocols.join(' ')}</span>
                    {token && <span style={{ color: '#e36209' }}> --token "{token}"</span>}
                  </Paper>
                </Box>
              </Stack>
            </TabPanel>

            {/* === Tab 2: 参数配置 === */}
            <TabPanel value={tabValue} index={1}>
              {configLoading ? (
                 <Box sx={{ display: 'flex', justifyContent: 'center', p: 5 }}>
                   <CircularProgress />
                 </Box>
              ) : (
                <Stack spacing={3}>
                  
                  {/* 1. 安全设置 */}
                  <Box>
                     <Typography variant="subtitle2" color="primary" sx={{ mb: 1.5, display: 'flex', alignItems: 'center', gap: 1 }}>
                        <VpnKeyIcon fontSize="small"/> 安全凭证
                     </Typography>
                     <Stack direction="row" spacing={1}>
                        <TextField 
                          fullWidth 
                          label="上报 Token" 
                          size="small" 
                          value={token} 
                          disabled 
                          // [修改点] 增加背景色，使其看起来像“暗色/不可修改”
                          sx={{ bgcolor: '#f5f5f5' }}
                          InputProps={{ endAdornment: (<InputAdornment position="end"></InputAdornment>) }}
                        />
                        <Button 
                          variant="contained" 
                          color="error" 
                          onClick={handleResetToken}
                          disabled={loading}
                          sx={{ minWidth: 80 }}
                        >
                          重置
                        </Button>
                     </Stack>
                  </Box>

                  {/* 2. 协议配置列表 */}
                  <Box>
                    <Typography variant="subtitle2" color="primary" sx={{ mb: 1.5, display: 'flex', alignItems: 'center', gap: 1 }}>
                       <SettingsIcon fontSize="small"/> 协议默认配置
                    </Typography>
                    
                    {PROTOCOL_CONFIGS.map((proto) => (
                      <Accordion 
                        key={proto.id} 
                        disableGutters 
                        elevation={0} 
                        // [修改点] 移除了 disabled={!proto.hasExtra}
                        // [修改点] 强制控制展开状态：如果没有额外配置，则锁定为 false (折叠)；否则为 undefined (由组件内部控制)
                        expanded={proto.hasExtra ? undefined : false}
                        sx={{ 
                          border: '1px solid #e0e0e0', 
                          '&:before': { display: 'none' }, 
                          mb: 1,
                          borderRadius: 1
                        }}
                      >
                        <AccordionSummary
                          // [修改点] 有配置显示展开箭头，无配置显示“减号”
                          expandIcon={
                            proto.hasExtra ? (
                              <ExpandMoreIcon />
                            ) : (
                              // 使用减号 (-) 作为占位，表示"该项固定/不可展开"
                              // color: 'action.disabled' 让它显示为浅灰色，不抢眼
                              <RemoveIcon sx={{ color: 'action.disabled', fontSize: '1.2rem' }} />
                            )
                          }
                          sx={{ 
                              bgcolor: '#fafafa', 
                              flexDirection: 'row-reverse',
                              cursor: proto.hasExtra ? 'pointer' : 'default',
                              '& .MuiAccordionSummary-content': { 
                                  alignItems: 'center', 
                                  justifyContent: 'space-between',
                                  ml: 1 
                              },
                              pointerEvents: proto.hasExtra ? 'auto' : 'none'
                          }}
                          onClick={(e) => !proto.hasExtra && e.stopPropagation()}
                        >
                          <Typography variant="body2" sx={{ fontWeight: 500 }}>{proto.name}</Typography>
                          
                          {/* ✅ 修改部分：右侧容器，包含标签和输入框 */}
                          <Stack 
                            direction="row" 
                            spacing={1} 
                            alignItems="center"
                            onClick={(e) => e.stopPropagation()} // 防止点击输入框触发折叠
                          >
                              {/* 渲染协议标签 */}
                              {proto.transports && proto.transports.map(type => (
                                <Chip 
                                  key={type}
                                  label={type} 
                                  size="small" 
                                  sx={{ 
                                    height: 20, 
                                    fontSize: '0.7rem', 
                                    fontWeight: 'bold',
                                    borderRadius: 1,
                                    // TCP 使用蓝色系，UDP 使用橙色系
                                    bgcolor: type === 'TCP' ? 'rgba(33, 150, 243, 0.1)' : 'rgba(255, 152, 0, 0.1)',
                                    color: type === 'TCP' ? '#1976d2' : '#ed6c02',
                                    border: '1px solid',
                                    borderColor: type === 'TCP' ? 'rgba(33, 150, 243, 0.3)' : 'rgba(255, 152, 0, 0.3)'
                                  }} 
                                />
                              ))}

                              <Box sx={{ width: '100px' }}>
                                  <TextField
                                      label="默认端口"
                                      size="small"
                                      variant="outlined"
                                      value={config[proto.portField]}
                                      onChange={handleConfigChange(proto.portField)}
                                      fullWidth
                                      sx={{ bgcolor: 'white' }}
                                  />
                              </Box>
                          </Stack>
                        </AccordionSummary>
                        
                        {proto.hasExtra && (
                           <AccordionDetails sx={{ borderTop: '1px solid #f0f0f0', bgcolor: '#fff' }}>
                              <Stack direction="row" spacing={2} flexWrap="wrap" useFlexGap>

                                 {proto.extraFields.map((field) => {
                                     if (field.type === 'toggle') {
                                       return (
                                          <Box 
                                            key={field.field} 
                                            sx={{ 
                                              flexGrow: field.fullWidth ? 1 : 0, 
                                              width: field.fullWidth ? '100%' : 'auto',
                                              minWidth: '45%'
                                            }}
                                          >
                                            <Typography variant="caption" color="textSecondary" sx={{ mb: 0.5, ml: 0.5, display: 'block' }}>
                                              {field.label}
                                            </Typography>
                                            <ToggleButtonGroup
                                              color="primary"
                                              value={config[field.field]}
                                              exclusive
                                              onChange={(e, newValue) => {
                                                if (newValue !== null) {
                                                  setConfig({ ...config, [field.field]: newValue });
                                                }
                                              }}
                                              fullWidth
                                              size="small"
                                            >
                                              {field.options.map((opt) => (
                                                <ToggleButton key={opt.value} value={opt.value}>
                                                  {opt.label}
                                                </ToggleButton>
                                              ))}
                                            </ToggleButtonGroup>
                                          </Box>
                                       );
                                     }

                                     return (
                                       <TextField
                                          key={field.field}
                                          label={field.label}
                                          value={config[field.field]}
                                          onChange={handleConfigChange(field.field)}
                                          size="small"
                                          helperText={field.helperText}
                                          sx={{ 
                                              flexGrow: field.fullWidth ? 1 : 0, 
                                              width: field.fullWidth ? '100%' : 'auto',
                                              minWidth: '45%'
                                          }}
                                       />
                                     );
                                   })}

                              </Stack>
                           </AccordionDetails>
                        )}

                      </Accordion>
                    ))}
                  </Box>

                  {/* 保存按钮 */}
                  <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <Button 
                      variant="contained" 
                      startIcon={loading ? <CircularProgress size={20} color="inherit"/> : <SaveIcon />}
                      onClick={handleSaveConfig}
                      disabled={loading}
                      fullWidth
                    >
                      保存所有配置
                    </Button>
                  </Box>
                </Stack>
              )}
            </TabPanel>

          </Box>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={onClose} color="inherit">关闭</Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert onClose={handleCloseSnackbar} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
};

AutoReportDialog.propTypes = {
  open: PropTypes.bool,
  onClose: PropTypes.func
};

export default AutoReportDialog;