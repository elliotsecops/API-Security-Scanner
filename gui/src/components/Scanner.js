import React, { useState } from 'react';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Box,
  Button,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip,
  LinearProgress,
  Alert,
  List,
  ListItem,
  ListItemText,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Tabs,
  Tab,
  Paper,
} from '@mui/material';
import {
  PlayArrow,
  Stop,
  Add,
  Delete,
  Visibility,
  Settings,
  Security,
  CheckCircle,
  Error,
  Warning,
} from '@mui/icons-material';

const Scanner = () => {
  const [scanConfig, setScanConfig] = useState({
    name: '',
    tenantId: '',
    endpoints: [{ url: '', method: 'GET', body: '', headers: {} }],
    payloads: {
      sql: ["' OR '1'='1", "'; DROP TABLE users;--"],
      xss: ["<script>alert('XSS')</script>", "'><script>alert('XSS')</script>"],
      nosql: ["{$ne: null}", "{$gt: ''}"],
    },
    rateLimiting: {
      requestsPerSecond: 10,
      maxConcurrent: 5,
    },
    options: {
      outputFormat: 'json',
      includeDetails: true,
      saveToHistory: true,
    },
  });

  const [activeScans, setActiveScans] = useState([]);
  const [scanHistory] = useState([]);
  const [selectedScan, setSelectedScan] = useState(null);
  const [resultsDialogOpen, setResultsDialogOpen] = useState(false);
  const [, setConfigDialogOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [tabValue, setTabValue] = useState(0);


  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  const addEndpoint = () => {
    setScanConfig({
      ...scanConfig,
      endpoints: [...scanConfig.endpoints, { url: '', method: 'GET', body: '', headers: {} }],
    });
  };

  const removeEndpoint = (index) => {
    const newEndpoints = scanConfig.endpoints.filter((_, i) => i !== index);
    setScanConfig({ ...scanConfig, endpoints: newEndpoints });
  };

  const updateEndpoint = (index, field, value) => {
    const newEndpoints = [...scanConfig.endpoints];
    newEndpoints[index][field] = value;
    setScanConfig({ ...scanConfig, endpoints: newEndpoints });
  };

  const startScan = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/scans', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(scanConfig),
      });

      if (!response.ok) {
        throw new Error('Failed to start scan');
      }

      const result = await response.json();
      setActiveScans([...activeScans, result.scan]);
    } catch (error) {
      console.error('Error starting scan:', error);
    } finally {
      setLoading(false);
    }
  };

  const stopScan = async (scanId) => {
    try {
      const response = await fetch(`/api/scans/${scanId}/stop`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (response.ok) {
        setActiveScans(activeScans.filter(scan => scan.id !== scanId));
      }
    } catch (error) {
      console.error('Error stopping scan:', error);
    }
  };

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'critical': return 'error';
      case 'high': return 'warning';
      case 'medium': return 'info';
      case 'low': return 'success';
      default: return 'default';
    }
  };

  const getSeverityIcon = (severity) => {
    switch (severity) {
      case 'critical': return <Error color="error" />;
      case 'high': return <Warning color="warning" />;
      case 'medium': return <Warning color="info" />;
      case 'low': return <CheckCircle color="success" />;
      default: return <Security />;
    }
  };

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Security Scanner
      </Typography>

      <Tabs value={tabValue} onChange={handleTabChange} sx={{ mb: 3 }}>
        <Tab label="Quick Scan" />
        <Tab label="Advanced Configuration" />
        <Tab label="Active Scans" />
        <Tab label="Scan History" />
      </Tabs>

      {/* Quick Scan Tab */}
      {tabValue === 0 && (
        <Grid container spacing={3}>
          <Grid item xs={12} md={8}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Quick Scan Configuration
                </Typography>

                <Grid container spacing={2}>
                  <Grid item xs={12}>
                    <TextField
                      fullWidth
                      label="Scan Name"
                      value={scanConfig.name}
                      onChange={(e) => setScanConfig({ ...scanConfig, name: e.target.value })}
                    />
                  </Grid>

                  <Grid item xs={12}>
                    <FormControl fullWidth>
                      <InputLabel>Tenant</InputLabel>
                      <Select
                        value={scanConfig.tenantId}
                        label="Tenant"
                        onChange={(e) => setScanConfig({ ...scanConfig, tenantId: e.target.value })}
                      >
                        <MenuItem value="tenant-001">Enterprise Corp</MenuItem>
                        <MenuItem value="tenant-002">Development Team</MenuItem>
                        <MenuItem value="tenant-003">Testing Environment</MenuItem>
                      </Select>
                    </FormControl>
                  </Grid>

                  <Grid item xs={12}>
                    <Typography variant="subtitle1" gutterBottom>
                      API Endpoints
                    </Typography>
                    {scanConfig.endpoints.map((endpoint, index) => (
                      <Paper key={index} sx={{ p: 2, mb: 2 }}>
                        <Grid container spacing={2}>
                          <Grid item xs={12} md={6}>
                            <TextField
                              fullWidth
                              label="URL"
                              value={endpoint.url}
                              onChange={(e) => updateEndpoint(index, 'url', e.target.value)}
                              placeholder="https://api.example.com/endpoint"
                            />
                          </Grid>
                          <Grid item xs={12} md={4}>
                            <FormControl fullWidth>
                              <InputLabel>Method</InputLabel>
                              <Select
                                value={endpoint.method}
                                label="Method"
                                onChange={(e) => updateEndpoint(index, 'method', e.target.value)}
                              >
                                <MenuItem value="GET">GET</MenuItem>
                                <MenuItem value="POST">POST</MenuItem>
                                <MenuItem value="PUT">PUT</MenuItem>
                                <MenuItem value="DELETE">DELETE</MenuItem>
                                <MenuItem value="PATCH">PATCH</MenuItem>
                              </Select>
                            </FormControl>
                          </Grid>
                          <Grid item xs={12} md={2}>
                            <IconButton
                              color="error"
                              onClick={() => removeEndpoint(index)}
                              disabled={scanConfig.endpoints.length === 1}
                            >
                              <Delete />
                            </IconButton>
                          </Grid>
                          {(endpoint.method === 'POST' || endpoint.method === 'PUT') && (
                            <Grid item xs={12}>
                              <TextField
                                fullWidth
                                label="Request Body"
                                value={endpoint.body}
                                onChange={(e) => updateEndpoint(index, 'body', e.target.value)}
                                multiline
                                rows={3}
                                placeholder='{"key": "value"}'
                              />
                            </Grid>
                          )}
                        </Grid>
                      </Paper>
                    ))}
                    <Button
                      startIcon={<Add />}
                      onClick={addEndpoint}
                      variant="outlined"
                    >
                      Add Endpoint
                    </Button>
                  </Grid>
                </Grid>

                <Box sx={{ mt: 3 }}>
                  <Button
                    variant="contained"
                    startIcon={<PlayArrow />}
                    onClick={startScan}
                    disabled={loading || !scanConfig.name || scanConfig.endpoints.some(e => !e.url)}
                    sx={{ mr: 2 }}
                  >
                    {loading ? <LinearProgress /> : 'Start Scan'}
                  </Button>
                  <Button
                    variant="outlined"
                    startIcon={<Settings />}
                    onClick={() => setConfigDialogOpen(true)}
                  >
                    Advanced Settings
                  </Button>
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} md={4}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Quick Templates
                </Typography>
                <List>
                  <ListItem button onClick={() => {
                    setScanConfig({
                      ...scanConfig,
                      name: 'REST API Scan',
                      endpoints: [
                        { url: 'https://api.example.com/users', method: 'GET' },
                        { url: 'https://api.example.com/data', method: 'POST', body: '{"query": "value"}' }
                      ]
                    });
                  }}>
                    <ListItemText
                      primary="REST API Assessment"
                      secondary="Scan typical REST API endpoints"
                    />
                  </ListItem>
                  <ListItem button onClick={() => {
                    setScanConfig({
                      ...scanConfig,
                      name: 'Auth Testing',
                      endpoints: [
                        { url: 'https://api.example.com/login', method: 'POST', body: '{"username": "test", "password": "test"}' },
                        { url: 'https://api.example.com/users', method: 'GET' }
                      ]
                    });
                  }}>
                    <ListItemText
                      primary="Authentication Testing"
                      secondary="Test auth bypass and session management"
                    />
                  </ListItem>
                  <ListItem button onClick={() => {
                    setScanConfig({
                      ...scanConfig,
                      name: 'Comprehensive Scan',
                      endpoints: [
                        { url: 'https://api.example.com/', method: 'GET' },
                        { url: 'https://api.example.com/api', method: 'GET' },
                        { url: 'https://api.example.com/v1/users', method: 'GET' },
                        { url: 'https://api.example.com/v1/data', method: 'POST', body: '{"test": "data"}' }
                      ]
                    });
                  }}>
                    <ListItemText
                      primary="Comprehensive Scan"
                      secondary="Full security assessment with multiple endpoints"
                    />
                  </ListItem>
                </List>
              </CardContent>
            </Card>

            <Card sx={{ mt: 2 }}>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Scan Statistics
                </Typography>
                <Grid container spacing={1}>
                  <Grid item xs={6}>
                    <Typography variant="body2">Total Scans:</Typography>
                    <Typography variant="h6">{scanHistory.length}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2">Active Scans:</Typography>
                    <Typography variant="h6">{activeScans.length}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2">Avg. Score:</Typography>
                    <Typography variant="h6">
                      {scanHistory.length > 0
                        ? Math.round(scanHistory.reduce((sum, scan) => sum + (scan.averageScore || 0), 0) / scanHistory.length)
                        : 0
                      }%
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2">Vulnerabilities:</Typography>
                    <Typography variant="h6">
                      {scanHistory.reduce((sum, scan) => sum + (scan.totalVulnerabilities || 0), 0)}
                    </Typography>
                  </Grid>
                </Grid>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {/* Active Scans Tab */}
      {tabValue === 2 && (
        <Grid container spacing={3}>
          {activeScans.map((scan) => (
            <Grid item xs={12} key={scan.id}>
              <Card>
                <CardContent>
                  <Grid container alignItems="center" justifyContent="space-between">
                    <Grid item>
                      <Typography variant="h6">{scan.name}</Typography>
                      <Typography variant="body2" color="textSecondary">
                        {scan.tenantId} • Started: {new Date(scan.startedAt).toLocaleString()}
                      </Typography>
                    </Grid>
                    <Grid item>
                      <Chip
                        label={scan.status}
                        color={scan.status === 'running' ? 'primary' : 'default'}
                        sx={{ mr: 1 }}
                      />
                      <Chip
                        label={`${scan.progress}%`}
                        variant="outlined"
                        sx={{ mr: 1 }}
                      />
                      <IconButton
                        color="error"
                        onClick={() => stopScan(scan.id)}
                        disabled={scan.status !== 'running'}
                      >
                        <Stop />
                      </IconButton>
                      <IconButton
                        onClick={() => {
                          setSelectedScan(scan);
                          setResultsDialogOpen(true);
                        }}
                      >
                        <Visibility />
                      </IconButton>
                    </Grid>
                  </Grid>
                  <LinearProgress
                    variant="determinate"
                    value={scan.progress}
                    sx={{ mt: 2 }}
                  />
                  <Typography variant="body2" sx={{ mt: 1 }}>
                    {scan.completedEndpoints}/{scan.totalEndpoints} endpoints completed
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
          {activeScans.length === 0 && (
            <Grid item xs={12}>
              <Alert severity="info">No active scans running</Alert>
            </Grid>
          )}
        </Grid>
      )}

      {/* Scan History Tab */}
      {tabValue === 3 && (
        <Grid container spacing={3}>
          {scanHistory.map((scan) => (
            <Grid item xs={12} key={scan.id}>
              <Card>
                <CardContent>
                  <Grid container alignItems="center" justifyContent="space-between">
                    <Grid item>
                      <Typography variant="h6">{scan.name}</Typography>
                      <Typography variant="body2" color="textSecondary">
                        {scan.tenantId} • {new Date(scan.completedAt).toLocaleString()}
                      </Typography>
                    </Grid>
                    <Grid item>
                      <Chip
                        icon={getSeverityIcon(scan.riskLevel)}
                        label={scan.riskLevel}
                        color={getSeverityColor(scan.riskLevel)}
                        sx={{ mr: 1 }}
                      />
                      <Chip
                        label={`${scan.averageScore}%`}
                        variant="outlined"
                        sx={{ mr: 1 }}
                      />
                      <IconButton onClick={() => {
                        setSelectedScan(scan);
                        setResultsDialogOpen(true);
                      }}>
                        <Visibility />
                      </IconButton>
                    </Grid>
                  </Grid>
                  <Box sx={{ mt: 2 }}>
                    <Typography variant="body2">
                      Endpoints: {scan.endpointsCount} • Vulnerabilities: {scan.totalVulnerabilities} • Duration: {Math.round(scan.duration / 60)}min
                    </Typography>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          ))}
          {scanHistory.length === 0 && (
            <Grid item xs={12}>
              <Alert severity="info">No scan history available</Alert>
            </Grid>
          )}
        </Grid>
      )}

      {/* Results Dialog */}
      <Dialog
        open={resultsDialogOpen}
        onClose={() => setResultsDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          Scan Results - {selectedScan?.name}
        </DialogTitle>
        <DialogContent>
          {selectedScan && (
            <Box>
              <Grid container spacing={2}>
                <Grid item xs={12} md={6}>
                  <Typography variant="subtitle1">Overview</Typography>
                  <Typography variant="body2">
                    Score: {selectedScan.averageScore}%
                  </Typography>
                  <Typography variant="body2">
                    Duration: {Math.round(selectedScan.duration / 60)} minutes
                  </Typography>
                  <Typography variant="body2">
                    Risk Level: {selectedScan.riskLevel}
                  </Typography>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="subtitle1">Vulnerabilities</Typography>
                  <Typography variant="body2">
                    Total: {selectedScan.totalVulnerabilities}
                  </Typography>
                  <Typography variant="body2">
                    Critical: {selectedScan.criticalVulnerabilities || 0}
                  </Typography>
                  <Typography variant="body2">
                    High: {selectedScan.highVulnerabilities || 0}
                  </Typography>
                </Grid>
              </Grid>

              <Typography variant="subtitle1" sx={{ mt: 2 }}>
                Endpoint Results
              </Typography>
              <List>
                {selectedScan.endpoints?.map((endpoint, index) => (
                  <ListItem key={index}>
                    <ListItemText
                      primary={endpoint.url}
                      secondary={`Score: ${endpoint.score}% • Vulnerabilities: ${endpoint.vulnerabilities || 0}`}
                    />
                  </ListItem>
                ))}
              </List>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setResultsDialogOpen(false)}>
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Scanner;