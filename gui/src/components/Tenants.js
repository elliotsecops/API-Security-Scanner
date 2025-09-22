import React, { useState, useEffect } from 'react';
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
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Alert,
  LinearProgress,
  Tabs,
  Tab,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Switch,
  FormControlLabel,
} from '@mui/material';
import {
  Add,
  Edit,
  Delete,
  Visibility,
  TrendingUp,
  TrendingDown,
  ExpandMore,
  People,
  Security,
  Storage,
  Speed,
} from '@mui/icons-material';
import { Chart as ChartJS, CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend, LineElement, PointElement } from 'chart.js';
import { Bar, Line } from 'react-chartjs-2';

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend, LineElement, PointElement);

const Tenants = () => {
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedTenant, setSelectedTenant] = useState(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [tabValue, setTabValue] = useState(0);
  const [tenantStats, setTenantStats] = useState(null);

  const [formData, setFormData] = useState({
    id: '',
    name: '',
    description: '',
    is_active: true,
    settings: {
      resource_limits: {
        max_requests_per_day: 10000,
        max_concurrent_scans: 5,
        max_endpoints_per_scan: 100,
        max_storage_mb: 1000,
      },
      data_isolation: {
        storage_path: './data/',
        enabled: true,
      },
      notification_settings: {
        email_notifications: true,
        email_recipients: [],
        alert_threshold: 'medium',
      },
    },
  });

  useEffect(() => {
    fetchTenants();
  }, []);

  const fetchTenants = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/tenants', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      const data = await response.json();
      setTenants(data.tenants || []);
    } catch (error) {
      console.error('Error fetching tenants:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchTenantStats = async (tenantId) => {
    try {
      const response = await fetch(`/api/tenants/${tenantId}/stats?period=30d`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      const data = await response.json();
      setTenantStats(data);
    } catch (error) {
      console.error('Error fetching tenant stats:', error);
    }
  };

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  const openCreateDialog = () => {
    setFormData({
      id: '',
      name: '',
      description: '',
      is_active: true,
      settings: {
        resource_limits: {
          max_requests_per_day: 10000,
          max_concurrent_scans: 5,
          max_endpoints_per_scan: 100,
          max_storage_mb: 1000,
        },
        data_isolation: {
          storage_path: './data/',
          enabled: true,
        },
        notification_settings: {
          email_notifications: true,
          email_recipients: [],
          alert_threshold: 'medium',
        },
      },
    });
    setEditMode(false);
    setDialogOpen(true);
  };

  const openEditDialog = (tenant) => {
    setFormData(tenant);
    setEditMode(true);
    setDialogOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const url = editMode ? `/api/tenants/${formData.id}` : '/api/tenants';
      const method = editMode ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        await fetchTenants();
        setDialogOpen(false);
      }
    } catch (error) {
      console.error('Error saving tenant:', error);
    }
  };

  const handleDelete = async (tenantId) => {
    if (window.confirm('Are you sure you want to delete this tenant?')) {
      try {
        const response = await fetch(`/api/tenants/${tenantId}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
          },
        });

        if (response.ok) {
          await fetchTenants();
        }
      } catch (error) {
        console.error('Error deleting tenant:', error);
      }
    }
  };

  const handleViewStats = (tenant) => {
    setSelectedTenant(tenant);
    fetchTenantStats(tenant.id);
    setTabValue(1);
  };

  // Chart data for tenant usage
  const usageData = {
    labels: tenants.map(t => t.name),
    datasets: [
      {
        label: 'Requests Used',
        data: tenants.map(t => t.settings?.resource_limits?.max_requests_per_day || 0),
        backgroundColor: 'rgba(25, 118, 210, 0.8)',
      },
      {
        label: 'Storage Used (MB)',
        data: tenants.map(t => t.settings?.resource_limits?.max_storage_mb || 0),
        backgroundColor: 'rgba(220, 53, 69, 0.8)',
      },
    ],
  };

  // Chart data for trends
  const trendData = tenantStats ? {
    labels: tenantStats.scan_trends?.map(t => new Date(t.date).toLocaleDateString()) || [],
    datasets: [
      {
        label: 'Vulnerabilities',
        data: tenantStats.scan_trends?.map(t => t.vulnerabilities) || [],
        borderColor: 'rgb(220, 53, 69)',
        backgroundColor: 'rgba(220, 53, 69, 0.2)',
        tension: 0.1,
      },
      {
        label: 'Average Score',
        data: tenantStats.scan_trends?.map(t => t.average_score) || [],
        borderColor: 'rgb(40, 167, 69)',
        backgroundColor: 'rgba(40, 167, 69, 0.2)',
        tension: 0.1,
      },
    ],
  } : null;

  if (loading) {
    return <LinearProgress />;
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Tenant Management
      </Typography>

      <Tabs value={tabValue} onChange={handleTabChange} sx={{ mb: 3 }}>
        <Tab label="Tenants" />
        <Tab label="Analytics" />
        <Tab label="Resource Usage" />
      </Tabs>

      {/* Tenants Tab */}
      {tabValue === 0 && (
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Typography variant="h6">
                    Tenant List ({tenants.length})
                  </Typography>
                  <Button
                    variant="contained"
                    startIcon={<Add />}
                    onClick={openCreateDialog}
                  >
                    Create Tenant
                  </Button>
                </Box>
              </CardContent>
            </Card>
          </Grid>

          {tenants.map((tenant) => (
            <Grid item xs={12} md={6} lg={4} key={tenant.id}>
              <Card>
                <CardContent>
                  <Box display="flex" justifyContent="space-between" alignItems="flex-start">
                    <Box>
                      <Typography variant="h6">
                        {tenant.name}
                      </Typography>
                      <Typography variant="body2" color="textSecondary">
                        {tenant.description}
                      </Typography>
                      <Chip
                        label={tenant.is_active ? 'Active' : 'Inactive'}
                        color={tenant.is_active ? 'success' : 'default'}
                        size="small"
                        sx={{ mt: 1 }}
                      />
                    </Box>
                    <Box>
                      <IconButton
                        size="small"
                        onClick={() => openEditDialog(tenant)}
                      >
                        <Edit />
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={() => handleViewStats(tenant)}
                      >
                        <Visibility />
                      </IconButton>
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => handleDelete(tenant.id)}
                      >
                        <Delete />
                      </IconButton>
                    </Box>
                  </Box>

                  <Box sx={{ mt: 2 }}>
                    <Typography variant="subtitle2">Resource Limits</Typography>
                    <Typography variant="body2">
                      <Speed fontSize="small" /> {tenant.settings?.resource_limits?.max_requests_per_day || 0} requests/day
                    </Typography>
                    <Typography variant="body2">
                      <People fontSize="small" /> {tenant.settings?.resource_limits?.max_concurrent_scans || 0} concurrent scans
                    </Typography>
                    <Typography variant="body2">
                      <Storage fontSize="small" /> {tenant.settings?.resource_limits?.max_storage_mb || 0} MB storage
                    </Typography>
                  </Box>

                  <Box sx={{ mt: 2 }}>
                    <Typography variant="subtitle2">Statistics</Typography>
                    <Typography variant="body2">
                      Total Scans: {tenant.stats?.total_scans || 0}
                    </Typography>
                    <Typography variant="body2">
                      Vulnerabilities: {tenant.stats?.vulnerabilities_found || 0}
                    </Typography>
                    <Typography variant="body2">
                      Average Score: {tenant.stats?.average_score || 0}%
                    </Typography>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Analytics Tab */}
      {tabValue === 1 && selectedTenant && (
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Analytics for {selectedTenant.name}
                </Typography>
                {tenantStats && (
                  <Grid container spacing={2}>
                    <Grid item xs={12} md={3}>
                      <Typography variant="body2">
                        <strong>Total Scans:</strong> {tenantStats.total_scans || 0}
                      </Typography>
                    </Grid>
                    <Grid item xs={12} md={3}>
                      <Typography variant="body2">
                        <strong>Successful:</strong> {tenantStats.successful_scans || 0}
                      </Typography>
                    </Grid>
                    <Grid item xs={12} md={3}>
                      <Typography variant="body2">
                        <strong>Failed:</strong> {tenantStats.failed_scans || 0}
                      </Typography>
                    </Grid>
                    <Grid item xs={12} md={3}>
                      <Typography variant="body2">
                        <strong>Success Rate:</strong> {tenantStats.total_scans ? Math.round((tenantStats.successful_scans / tenantStats.total_scans) * 100) : 0}%
                      </Typography>
                    </Grid>
                  </Grid>
                )}
              </CardContent>
            </Card>
          </Grid>

          {trendData && (
            <Grid item xs={12}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Trends (Last 30 Days)
                  </Typography>
                  <Line data={trendData} options={{ responsive: true }} />
                </CardContent>
              </Card>
            </Grid>
          )}

          {tenantStats && (
            <Grid item xs={12}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Resource Usage
                  </Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={12} md={4}>
                      <Typography variant="body2">
                        <strong>Requests:</strong> {tenantStats.resource_usage?.requests_used || 0} / {tenantStats.resource_usage?.requests_limit || 0}
                      </Typography>
                      <LinearProgress
                        variant="determinate"
                        value={tenantStats.resource_usage?.requests_limit ? (tenantStats.resource_usage.requests_used / tenantStats.resource_usage.requests_limit) * 100 : 0}
                        sx={{ mt: 1 }}
                      />
                    </Grid>
                    <Grid item xs={12} md={4}>
                      <Typography variant="body2">
                        <strong>Storage:</strong> {tenantStats.resource_usage?.storage_used || 0} / {tenantStats.resource_usage?.storage_limit || 0} MB
                      </Typography>
                      <LinearProgress
                        variant="determinate"
                        value={tenantStats.resource_usage?.storage_limit ? (tenantStats.resource_usage.storage_used / tenantStats.resource_usage.storage_limit) * 100 : 0}
                        sx={{ mt: 1 }}
                      />
                    </Grid>
                    <Grid item xs={12} md={4}>
                      <Typography variant="body2">
                        <strong>Concurrent Scans:</strong> {tenantStats.resource_usage?.concurrent_scans || 0} / {tenantStats.resource_usage?.concurrent_limit || 0}
                      </Typography>
                      <LinearProgress
                        variant="determinate"
                        value={tenantStats.resource_usage?.concurrent_limit ? (tenantStats.resource_usage.concurrent_scans / tenantStats.resource_usage.concurrent_limit) * 100 : 0}
                        sx={{ mt: 1 }}
                      />
                    </Grid>
                  </Grid>
                </CardContent>
              </Card>
            </Grid>
          )}
        </Grid>
      )}

      {/* Resource Usage Tab */}
      {tabValue === 2 && (
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Resource Usage Across All Tenants
                </Typography>
                <Bar data={usageData} options={{ responsive: true }} />
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Resource Allocation Summary
                </Typography>
                <TableContainer component={Paper}>
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Tenant</TableCell>
                        <TableCell>Status</TableCell>
                        <TableCell>Requests/Day</TableCell>
                        <TableCell>Storage (MB)</TableCell>
                        <TableCell>Concurrent Scans</TableCell>
                        <TableCell>Usage</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {tenants.map((tenant) => {
                        const usage = tenant.stats?.average_score || 0;
                        return (
                          <TableRow key={tenant.id}>
                            <TableCell>{tenant.name}</TableCell>
                            <TableCell>
                              <Chip
                                label={tenant.is_active ? 'Active' : 'Inactive'}
                                color={tenant.is_active ? 'success' : 'default'}
                                size="small"
                              />
                            </TableCell>
                            <TableCell>{tenant.settings?.resource_limits?.max_requests_per_day || 0}</TableCell>
                            <TableCell>{tenant.settings?.resource_limits?.max_storage_mb || 0}</TableCell>
                            <TableCell>{tenant.settings?.resource_limits?.max_concurrent_scans || 0}</TableCell>
                            <TableCell>
                              <LinearProgress
                                variant="determinate"
                                value={usage}
                                sx={{ width: 100 }}
                              />
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </TableContainer>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {/* Tenant Dialog */}
      <Dialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          {editMode ? 'Edit Tenant' : 'Create New Tenant'}
        </DialogTitle>
        <DialogContent>
          <Grid container spacing={2}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Tenant ID"
                value={formData.id}
                onChange={(e) => setFormData({ ...formData, id: e.target.value })}
                disabled={editMode}
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                multiline
                rows={3}
              />
            </Grid>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                  />
                }
                label="Active"
              />
            </Grid>

            {/* Resource Limits */}
            <Grid item xs={12}>
              <Accordion>
                <AccordionSummary expandIcon={<ExpandMore />}>
                  <Typography>Resource Limits</Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <Grid container spacing={2}>
                    <Grid item xs={12} md={4}>
                      <TextField
                        fullWidth
                        label="Max Requests/Day"
                        type="number"
                        value={formData.settings.resource_limits.max_requests_per_day}
                        onChange={(e) => setFormData({
                          ...formData,
                          settings: {
                            ...formData.settings,
                            resource_limits: {
                              ...formData.settings.resource_limits,
                              max_requests_per_day: parseInt(e.target.value)
                            }
                          }
                        })}
                      />
                    </Grid>
                    <Grid item xs={12} md={4}>
                      <TextField
                        fullWidth
                        label="Max Concurrent Scans"
                        type="number"
                        value={formData.settings.resource_limits.max_concurrent_scans}
                        onChange={(e) => setFormData({
                          ...formData,
                          settings: {
                            ...formData.settings,
                            resource_limits: {
                              ...formData.settings.resource_limits,
                              max_concurrent_scans: parseInt(e.target.value)
                            }
                          }
                        })}
                      />
                    </Grid>
                    <Grid item xs={12} md={4}>
                      <TextField
                        fullWidth
                        label="Max Storage (MB)"
                        type="number"
                        value={formData.settings.resource_limits.max_storage_mb}
                        onChange={(e) => setFormData({
                          ...formData,
                          settings: {
                            ...formData.settings,
                            resource_limits: {
                              ...formData.settings.resource_limits,
                              max_storage_mb: parseInt(e.target.value)
                            }
                          }
                        })}
                      />
                    </Grid>
                  </Grid>
                </AccordionDetails>
              </Accordion>
            </Grid>

            {/* Notification Settings */}
            <Grid item xs={12}>
              <Accordion>
                <AccordionSummary expandIcon={<ExpandMore />}>
                  <Typography>Notification Settings</Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <Grid container spacing={2}>
                    <Grid item xs={12}>
                      <FormControlLabel
                        control={
                          <Switch
                            checked={formData.settings.notification_settings.email_notifications}
                            onChange={(e) => setFormData({
                              ...formData,
                              settings: {
                                ...formData.settings,
                                notification_settings: {
                                  ...formData.settings.notification_settings,
                                  email_notifications: e.target.checked
                                }
                              }
                            })}
                          />
                        }
                        label="Email Notifications"
                      />
                    </Grid>
                    <Grid item xs={12}>
                      <FormControl fullWidth>
                        <InputLabel>Alert Threshold</InputLabel>
                        <Select
                          value={formData.settings.notification_settings.alert_threshold}
                          label="Alert Threshold"
                          onChange={(e) => setFormData({
                            ...formData,
                            settings: {
                              ...formData.settings,
                              notification_settings: {
                                ...formData.settings.notification_settings,
                                alert_threshold: e.target.value
                              }
                            }
                          })}
                        >
                          <MenuItem value="low">Low</MenuItem>
                          <MenuItem value="medium">Medium</MenuItem>
                          <MenuItem value="high">High</MenuItem>
                          <MenuItem value="critical">Critical</MenuItem>
                        </Select>
                      </FormControl>
                    </Grid>
                  </Grid>
                </AccordionDetails>
              </Accordion>
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} variant="contained">
            {editMode ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Tenants;