import React, { useState, useEffect } from 'react';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Box,
  LinearProgress,
  Button,
  Chip,
  Alert,
  List,
  ListItem,
  ListItemText,
  Divider,
} from '@mui/material';
import {
  Security,
  TrendingUp,
  TrendingDown,
  Warning,
  CheckCircle,
  Error,
  Schedule,
  Speed,
} from '@mui/icons-material';
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, BarElement, Title, Tooltip, Legend, ArcElement } from 'chart.js';
import { Line, Bar, Doughnut } from 'react-chartjs-2';
import { useMetrics } from '../contexts/MetricsContext';
import { useWebSocket } from '../contexts/WebSocketContext';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement
);

const Dashboard = () => {
  const { metrics, loading: metricsLoading, error: metricsError } = useMetrics();
  const { lastMessage, connectionStatus } = useWebSocket();
  const [systemHealth, setSystemHealth] = useState({});

  useEffect(() => {
    if (lastMessage && lastMessage.type === 'metrics') {
      // Update real-time metrics
      console.log('Real-time metrics update:', lastMessage.data);
    }
  }, [lastMessage]);

  useEffect(() => {
    // Fetch system health
    fetchSystemHealth();
  }, []);

  const fetchSystemHealth = async () => {
    try {
      const response = await fetch('/api/health');
      const data = await response.json();
      setSystemHealth(data);
    } catch (error) {
      console.error('Error fetching system health:', error);
    }
  };

  // Mock data for charts
  const vulnerabilityData = {
    labels: ['SQL Injection', 'XSS', 'Auth Bypass', 'Parameter Tampering', 'Header Issues'],
    datasets: [
      {
        label: 'Critical',
        data: [3, 2, 1, 0, 0],
        backgroundColor: 'rgba(220, 53, 69, 0.8)',
      },
      {
        label: 'High',
        data: [5, 8, 3, 2, 1],
        backgroundColor: 'rgba(255, 193, 7, 0.8)',
      },
      {
        label: 'Medium',
        data: [2, 3, 4, 5, 3],
        backgroundColor: 'rgba(255, 152, 0, 0.8)',
      },
    ],
  };

  const scanTrendData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Scans Completed',
        data: [12, 19, 15, 25, 22, 18, 20],
        borderColor: 'rgb(75, 192, 192)',
        backgroundColor: 'rgba(75, 192, 192, 0.2)',
        tension: 0.1,
      },
      {
        label: 'Vulnerabilities Found',
        data: [5, 8, 6, 12, 9, 7, 8],
        borderColor: 'rgb(255, 99, 132)',
        backgroundColor: 'rgba(255, 99, 132, 0.2)',
        tension: 0.1,
      },
    ],
  };

  const riskDistributionData = {
    labels: ['Critical', 'High', 'Medium', 'Low'],
    datasets: [
      {
        data: [6, 18, 12, 4],
        backgroundColor: [
          'rgba(220, 53, 69, 0.8)',
          'rgba(255, 193, 7, 0.8)',
          'rgba(255, 152, 0, 0.8)',
          'rgba(40, 167, 69, 0.8)',
        ],
        borderWidth: 0,
      },
    ],
  };

  const statsCards = [
    {
      title: 'Active Scans',
      value: metrics?.scanner?.active_scans || 0,
      icon: <Security sx={{ fontSize: 40 }} />,
      color: '#1976d2',
      trend: 'up',
    },
    {
      title: 'Vulnerabilities Found',
      value: metrics?.security?.vulnerabilities_found || 0,
      icon: <Warning sx={{ fontSize: 40 }} />,
      color: '#dc004e',
      trend: 'up',
    },
    {
      title: 'System Health',
      value: systemHealth.status === 'healthy' ? 'Healthy' : 'Issues',
      icon: systemHealth.status === 'healthy' ? <CheckCircle sx={{ fontSize: 40 }} /> : <Error sx={{ fontSize: 40 }} />,
      color: systemHealth.status === 'healthy' ? '#4caf50' : '#f44336',
    },
    {
      title: 'Response Time',
      value: `${metrics?.scanner?.average_response_time || 0}ms`,
      icon: <Speed sx={{ fontSize: 40 }} />,
      color: '#ff9800',
      trend: 'down',
    },
  ];

  if (metricsLoading) {
    return (
      <Box>
        <Typography variant="h4" gutterBottom>
          Dashboard
        </Typography>
        <LinearProgress />
      </Box>
    );
  }

  if (metricsError) {
    return (
      <Box>
        <Alert severity="error">
          Error loading metrics: {metricsError}
        </Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Security Dashboard
      </Typography>

      {/* Connection Status */}
      <Alert severity={connectionStatus === 'connected' ? 'success' : 'warning'} sx={{ mb: 2 }}>
        WebSocket Status: {connectionStatus}
      </Alert>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        {statsCards.map((stat, index) => (
          <Grid item xs={12} sm={6} md={3} key={index}>
            <Card sx={{ height: '100%' }}>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Box>
                    <Typography color="textSecondary" gutterBottom variant="h6">
                      {stat.title}
                    </Typography>
                    <Typography variant="h4" component="h2">
                      {stat.value}
                    </Typography>
                  </Box>
                  <Box sx={{ color: stat.color }}>
                    {stat.icon}
                  </Box>
                </Box>
                {stat.trend && (
                  <Box display="flex" alignItems="center" mt={2}>
                    {stat.trend === 'up' ? <TrendingUp /> : <TrendingDown />}
                    <Typography variant="body2" sx={{ ml: 1 }}>
                      {stat.trend === 'up' ? 'Increasing' : 'Decreasing'}
                    </Typography>
                  </Box>
                )}
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Charts */}
      <Grid container spacing={3}>
        {/* Vulnerability Chart */}
        <Grid item xs={12} md={8}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Vulnerability Distribution by Type
              </Typography>
              <Bar data={vulnerabilityData} options={{ responsive: true }} />
            </CardContent>
          </Card>
        </Grid>

        {/* Risk Distribution */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Risk Distribution
              </Typography>
              <Doughnut data={riskDistributionData} options={{ responsive: true }} />
            </CardContent>
          </Card>
        </Grid>

        {/* Scan Trends */}
        <Grid item xs={12} md={8}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Weekly Scan Trends
              </Typography>
              <Line data={scanTrendData} options={{ responsive: true }} />
            </CardContent>
          </Card>
        </Grid>

        {/* Recent Activity */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Recent Activity
              </Typography>
              <List dense>
                <ListItem>
                  <ListItemText
                    primary="Scan completed"
                    secondary="API endpoints scanned - 2 vulnerabilities found"
                  />
                </ListItem>
                <Divider />
                <ListItem>
                  <ListItemText
                    primary="New tenant created"
                    secondary="Enterprise Corp - Tenant ID: enterprise-001"
                  />
                </ListItem>
                <Divider />
                <ListItem>
                  <ListItemText
                    primary="SIEM integration test"
                    secondary="Wazuh connection test successful"
                  />
                </ListItem>
                <Divider />
                <ListItem>
                  <ListItemText
                    primary="Configuration updated"
                    secondary="Rate limiting settings modified"
                  />
                </ListItem>
              </List>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* System Information */}
      <Grid container spacing={3} sx={{ mt: 3 }}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                System Information
              </Typography>
              <Grid container spacing={2}>
                <Grid item xs={12} md={6}>
                  <Typography variant="body2">
                    <strong>CPU Usage:</strong> {metrics?.system?.cpu_usage || 0}%
                  </Typography>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="body2">
                    <strong>Memory Usage:</strong> {metrics?.system?.memory_percent || 0}%
                  </Typography>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="body2">
                    <strong>Active Tenants:</strong> {metrics?.tenants?.active_tenants || 0}
                  </Typography>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="body2">
                    <strong>Uptime:</strong> {Math.floor((metrics?.system?.uptime || 0) / 3600)}h
                  </Typography>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard;