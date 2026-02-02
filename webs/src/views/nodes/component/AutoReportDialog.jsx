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
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Tooltip from '@mui/material/Tooltip';
import Divider from '@mui/material/Divider';

// icons
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import CheckIcon from '@mui/icons-material/Check';
import TerminalIcon from '@mui/icons-material/Terminal';

// api
import { getSystemDomain } from 'api/settings';

const PROTOCOL_OPTIONS = [
  { label: 'VLESS', value: 'vless' },
  { label: 'Shadowsocks', value: 'shadowsocks' },
  { label: 'Hysteria2', value: 'hysteria2' },
  { label: 'Tuic', value: 'tuic' },
  { label: 'Socks5', value: 'socks5' }
];

const AutoReportDialog = ({ open, onClose }) => {
  const [baseUrl, setBaseUrl] = useState(window.location.origin);
  const [token, setToken] = useState('');
  const [selectedProtocols, setSelectedProtocols] = useState(['vless', 'shadowsocks', 'hysteria2']);
  const [copySuccess, setCopySuccess] = useState(false);

  // 监听打开状态，自动获取“远程访问域名”
  useEffect(() => {
    if (open) {
      getSystemDomain()
        .then((res) => {
          if (res.data && res.data.systemDomain) {
            let domain = res.data.systemDomain;
            if (!domain.startsWith('http')) {
              domain = `https://${domain}`;
            }
            setBaseUrl(domain);
          }
        })
        .catch((err) => {
          console.error("自动获取远程域名失败，将使用默认浏览器地址", err);
        });
    }
  }, [open]);

  // 核心逻辑：实时计算生成的命令
  const generatedCommand = useMemo(() => {
    const host = baseUrl.replace(/\/$/, '');
    const protocolStr = selectedProtocols.length > 0 ? selectedProtocols.join(' ') : 'vless';
    const tokenStr = token ? `--token "${token}"` : '';

    return `bash -c "$(curl -fsSL ${host}/subscription/install-singbox.sh)" -- ${protocolStr} --report "${host}/api/v1/nodes/report" ${tokenStr}`;
  }, [baseUrl, token, selectedProtocols]);

  const handleProtocolChange = (value) => {
    const currentIndex = selectedProtocols.indexOf(value);
    const newChecked = [...selectedProtocols];

    if (currentIndex === -1) {
      newChecked.push(value);
    } else {
      newChecked.splice(currentIndex, 1);
    }
    setSelectedProtocols(newChecked);
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(generatedCommand);
    setCopySuccess(true);
    setTimeout(() => setCopySuccess(false), 2000);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <TerminalIcon color="primary" />
        节点自动上报配置
      </DialogTitle>
      
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          <Alert severity="info" variant="outlined" sx={{ border: 'none', bgcolor: 'primary.lighter' }}>
            在您的节点服务器上执行下方命令，即可一键安装 sing-box 并自动上报节点信息到本面板。
          </Alert>

          {/* 1. 设置区域 */}
          <Box sx={{ p: 2, border: '1px solid', borderColor: 'divider', borderRadius: 2 }}>
            <Stack spacing={2}>
              <Typography variant="subtitle2" color="primary">参数设置</Typography>
              
              <TextField
                label="上报验证 Token (可选)"
                fullWidth
                size="small"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="建议设置密钥以防止恶意上报"
                helperText="如果不设置，任何知道接口地址的人都可以上报节点"
              />

              <Box>
                <Typography variant="caption" color="textSecondary" sx={{ mb: 1, display: 'block' }}>
                  安装协议选择
                </Typography>
                <FormGroup row sx={{ ml: 1 }}>
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
            </Stack>
          </Box>

          {/* 2. 命令展示区域 (美化版) */}
          <Box>
            <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1 }}>
              <Typography variant="subtitle2" color="primary">
                一键安装命令
              </Typography>
              <Typography variant="caption" color="textSecondary">
                上报至: {baseUrl}
              </Typography>
            </Stack>
            
            <Paper
              elevation={0}
              sx={{
                position: 'relative',
                bgcolor: '#1e1e1e', // 深色终端背景
                color: '#a9b7c6',   // 浅色代码文字
                p: 2,
                borderRadius: 2,
                fontFamily: '"JetBrains Mono", "Consolas", "Monaco", monospace', // 等宽字体
                fontSize: '0.85rem',
                lineHeight: 1.6,
                wordBreak: 'break-all',
                cursor: 'text',
                border: '1px solid #333',
                '&:hover .copy-btn': { opacity: 1 }
              }}
              onClick={(e) => {
                // 点击自动全选文本
                const range = document.createRange();
                range.selectNodeContents(e.currentTarget);
                const selection = window.getSelection();
                selection.removeAllRanges();
                selection.addRange(range);
              }}
            >
              {/* 复制按钮 */}
              <Tooltip title={copySuccess ? "已复制!" : "复制命令"}>
                <IconButton
                  className="copy-btn"
                  onClick={(e) => {
                    e.stopPropagation(); // 防止触发全选
                    handleCopy();
                  }}
                  size="small"
                  sx={{
                    position: 'absolute',
                    top: 8,
                    right: 8,
                    color: copySuccess ? '#4caf50' : '#fff',
                    bgcolor: 'rgba(255,255,255,0.1)',
                    opacity: 0.7,
                    transition: 'all 0.2s',
                    '&:hover': {
                      bgcolor: 'rgba(255,255,255,0.2)',
                      opacity: 1
                    }
                  }}
                >
                  {copySuccess ? <CheckIcon fontSize="small" /> : <ContentCopyIcon fontSize="small" />}
                </IconButton>
              </Tooltip>

              {/* 简单的语法高亮模拟 */}
              <span style={{ color: '#cc7832' }}>bash</span> -c <span style={{ color: '#6a8759' }}>"$(curl -fsSL {baseUrl}/subscription/install-singbox.sh)"</span> -- 
              <span style={{ color: '#9876aa' }}> {selectedProtocols.join(' ')}</span> --report <span style={{ color: '#6a8759' }}>"{baseUrl}/api/v1/nodes/report"</span>
              {token && <span style={{ color: '#ffc66d' }}> --token "{token}"</span>}
            </Paper>
            <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block', textAlign: 'right' }}>
              * 点击代码块可全选，点击右上角按钮复制
            </Typography>
          </Box>

        </Stack>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} color="inherit">关闭</Button>
      </DialogActions>
    </Dialog>
  );
};

AutoReportDialog.propTypes = {
  open: PropTypes.bool,
  onClose: PropTypes.func
};

export default AutoReportDialog;