import React, { useState, useEffect } from 'react';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Box,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Button,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Alert,
  LinearProgress,
} from '@mui/material';
import {
  GetApp,
  Visibility,
  FilterList,
  ExpandMore,
  Search,
  Security,
  Error,
  Warning,
  Info,
  CheckCircle,
} from '@mui/icons-material';
import { Chart as ChartJS, CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend, ArcElement } from 'chart.js';
import { Bar, Doughnut } from 'react-chartjs-2';

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend, ArcElement);

const Results = () => {
  const [scanResults, setScanResults] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedResult, setSelectedResult] = useState(null);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [filters, setFilters] = useState({
    tenant: '',
    severity: '',
    dateRange: '7d',
    search: '',
  });

  useEffect(() => {
    fetchScanResults();
  }, []);

  const fetchScanResults = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/scans', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      const data = await response.json();
      setScanResults(data.scans || []);
    } catch (error) {
      console.error('Error fetching scan results:', error);
    } finally {
      setLoading(false);
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
      case 'medium': return <Info color="info" />;
      case 'low': return <CheckCircle color="success" />;
      default: return <Security />;
    }
  };

  const filteredResults = scanResults.filter(result => {
    if (filters.tenant && result.tenant_id !== filters.tenant) return false;
    if (filters.severity && result.risk_level !== filters.severity) return false;
    if (filters.search && !result.name.toLowerCase().includes(filters.search.toLowerCase())) return false;
    return true;
  });

  // Vulnerability distribution data
  const vulnerabilityData = {
    labels: ['Critical', 'High', 'Medium', 'Low'],
    datasets: [
      {
        label: 'Vulnerabilities',
        data: [
          scanResults.reduce((sum, r) => sum + (r.critical_vulnerabilities || 0), 0),
          scanResults.reduce((sum, r) => sum + (r.high_vulnerabilities || 0), 0),
          scanResults.reduce((sum, r) => sum + (r.medium_vulnerabilities || 0), 0),
          scanResults.reduce((sum, r) => sum + (r.low_vulnerabilities || 0), 0),
        ],
        backgroundColor: [
          'rgba(220, 53, 69, 0.8)',
          'rgba(255, 193, 7, 0.8)',
          'rgba(255, 152, 0, 0.8)',
          'rgba(40, 167, 69, 0.8)',
        ],
      },
    ],
  };

  // Score trends data
  const scoreTrendsData = {
    labels: filteredResults.slice(-10).map(r => new Date(r.completed_at).toLocaleDateString()),
    datasets: [
      {
        label: 'Security Score',
        data: filteredResults.slice(-10).map(r => r.average_score),
        backgroundColor: 'rgba(25, 118, 210, 0.8)',
      },
    ],
  };

  if (loading) {
    return <LinearProgress />;
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Scan Results
      </Typography>

      {/* Summary Statistics */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Total Scans
              </Typography>
              <Typography variant="h4">
                {scanResults.length}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Total Vulnerabilities
              </Typography>
              <Typography variant="h4">
                {scanResults.reduce((sum, r) => sum + (r.total_vulnerabilities || 0), 0)}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Average Score
              </Typography>
              <Typography variant="h4">
                {scanResults.length > 0
                  ? Math.round(scanResults.reduce((sum, r) => sum + (r.average_score || 0), 0) / scanResults.length)
                  : 0}%
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Critical Issues
              </Typography>
              <Typography variant="h4">
                {scanResults.reduce((sum, r) => sum + (r.critical_vulnerabilities || 0), 0)}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Charts */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Vulnerability Distribution
              </Typography>
              <Doughnut data={vulnerabilityData} options={{ responsive: true }} />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Score Trends (Last 10 Scans)
              </Typography>
              <Bar data={scoreTrendsData} options={{ responsive: true }} />
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Filters */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Filters
          </Typography>
          <Grid container spacing={2}>
            <Grid item xs={12} md={3}>
              <FormControl fullWidth>
                <InputLabel>Tenant</InputLabel>
                <Select
                  value={filters.tenant}
                  label="Tenant"
                  onChange={(e) => setFilters({ ...filters, tenant: e.target.value })}
                >
                  <MenuItem value="">All Tenants</MenuItem>
                  <MenuItem value="tenant-001">Enterprise Corp</MenuItem>
                  <MenuItem value="tenant-002">Development Team</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={3}>
              <FormControl fullWidth>
                <InputLabel>Severity</InputLabel>
                <Select
                  value={filters.severity}
                  label="Severity"
                  onChange={(e) => setFilters({ ...filters, severity: e.target.value })}
                >
                  <MenuItem value="">All Severities</MenuItem>
                  <MenuItem value="critical">Critical</MenuItem>
                  <MenuItem value="high">High</MenuItem>
                  <MenuItem value="medium">Medium</MenuItem>
                  <MenuItem value="low">Low</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={3}>
              <FormControl fullWidth>
                <InputLabel>Date Range</InputLabel>
                <Select
                  value={filters.dateRange}
                  label="Date Range"
                  onChange={(e) => setFilters({ ...filters, dateRange: e.target.value })}
                >
                  <MenuItem value="1d">Last 24 Hours</MenuItem>
                  <MenuItem value="7d">Last 7 Days</MenuItem>
                  <MenuItem value="30d">Last 30 Days</MenuItem>
                  <MenuItem value="90d">Last 90 Days</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={3}>
              <TextField
                fullWidth
                label="Search"
                value={filters.search}
                onChange={(e) => setFilters({ ...filters, search: e.target.value })}
                InputProps={{ startAdornment: <Search /> }}
              />
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Results Table */}
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Scan Results ({filteredResults.length} found)
          </Typography>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Scan Name</TableCell>
                  <TableCell>Tenant</TableCell>
                  <TableCell>Date</TableCell>
                  <TableCell>Score</TableCell>
                  <TableCell>Risk Level</TableCell>
                  <TableCell>Vulnerabilities</TableCell>
                  <TableCell>Duration</TableCell>
                  <TableCell>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredResults.map((result) => (
                  <TableRow key={result.id}>
                    <TableCell>{result.name}</TableCell>
                    <TableCell>{result.tenant_id}</TableCell>
                    <TableCell>
                      {new Date(result.completed_at || result.started_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={`${result.average_score || 0}%`}
                        color={result.average_score >= 80 ? 'success' : result.average_score >= 60 ? 'warning' : 'error'}
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        icon={getSeverityIcon(result.risk_level)}
                        label={result.risk_level}
                        color={getSeverityColor(result.risk_level)}
                      />
                    </TableCell>
                    <TableCell>{result.total_vulnerabilities || 0}</TableCell>
                    <TableCell>
                      {result.duration ? `${Math.round(result.duration / 60)}min` : 'N/A'}
                    </TableCell>
                    <TableCell>
                      <Button
                        size="small"
                        onClick={() => {
                          setSelectedResult(result);
                          setDetailDialogOpen(true);
                        }}
                      >
                        <Visibility />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>

      {/* Detail Dialog */}
      <Dialog
        open={detailDialogOpen}
        onClose={() => setDetailDialogOpen(false)}
        maxWidth="lg"
        fullWidth
      >
        <DialogTitle>
          Scan Details - {selectedResult?.name}
        </DialogTitle>
        <DialogContent>
          {selectedResult && (
            <Box>
              <Grid container spacing={2} sx={{ mb: 2 }}>
                <Grid item xs={12} md={6}>
                  <Typography variant="subtitle1">General Information</Typography>
                  <Typography variant="body2">
                    <strong>Scan ID:</strong> {selectedResult.id}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Tenant:</strong> {selectedResult.tenant_id}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Started:</strong> {new Date(selectedResult.started_at).toLocaleString()}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Completed:</strong> {selectedResult.completed_at ? new Date(selectedResult.completed_at).toLocaleString() : 'N/A'}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Duration:</strong> {selectedResult.duration ? `${Math.round(selectedResult.duration / 60)} minutes` : 'N/A'}
                  </Typography>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="subtitle1">Security Assessment</Typography>
                  <Typography variant="body2">
                    <strong>Overall Score:</strong> {selectedResult.average_score || 0}%
                  </Typography>
                  <Typography variant="body2">
                    <strong>Risk Level:</strong> {selectedResult.risk_level}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Total Vulnerabilities:</strong> {selectedResult.total_vulnerabilities || 0}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Critical:</strong> {selectedResult.critical_vulnerabilities || 0}
                  </Typography>
                  <Typography variant="body2">
                    <strong>High:</strong> {selectedResult.high_vulnerabilities || 0}
                  </Typography>
                </Grid>
              </Grid>

              <Typography variant="subtitle1" gutterBottom>
                Endpoint Results
              </Typography>
              {selectedResult.endpoints?.map((endpoint, index) => (
                <Accordion key={index}>
                  <AccordionSummary expandIcon={<ExpandMore />}>
                    <Box display="flex" alignItems="center" width="100%">
                      <Typography sx={{ flexGrow: 1 }}>
                        {endpoint.url}
                      </Typography>
                      <Chip
                        label={`${endpoint.score || 0}%`}
                        color={endpoint.score >= 80 ? 'success' : endpoint.score >= 60 ? 'warning' : 'error'}
                        sx={{ mr: 1 }}
                      />
                      <Chip
                        label={`${endpoint.vulnerabilities || 0} vulns`}
                        variant="outlined"
                      />
                    </Box>
                  </AccordionSummary>
                  <AccordionDetails>
                    <Box width="100%">
                      <Typography variant="body2">
                        <strong>Method:</strong> {endpoint.method}
                      </Typography>
                      <Typography variant="body2">
                        <strong>Status:</strong> {endpoint.status}
                      </Typography>
                      {endpoint.results && (
                        <Box mt={2}>
                          <Typography variant="subtitle2">Test Results:</Typography>
                          {endpoint.results.map((test, testIndex) => (
                            <Alert
                              key={testIndex}
                              severity={test.passed ? 'success' : 'error'}
                              sx={{ mt: 1 }}
                            >
                              <Typography variant="body2">
                                <strong>{test.test_name}:</strong> {test.message}
                              </Typography>
                            </Alert>
                          ))}
                        </Box>
                      )}
                    </Box>
                  </AccordionDetails>
                </Accordion>
              ))}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailDialogOpen(false)}>
            Close
          </Button>
          <Button
            variant="contained"
            startIcon={<GetApp />}
            onClick={() => {
              // Download functionality would go here
            }}
          >
            Export Report
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Results;