import React, { useMemo, useState } from 'react';
import {
  Box,
  Typography,
  Paper,
  Grid,
  TextField,
  Button,
  Alert,
  Divider,
  useTheme,
} from '@mui/material';
import LockResetIcon from '@mui/icons-material/LockReset';
import { useAuth } from '../contexts/AuthContext';

export default function Settings() {
  const theme = useTheme();
  const { user, token, login } = useAuth();
  const [form, setForm] = useState({
    currentUsername: user?.username || 'admin',
    newUsername: user?.username || 'admin',
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [feedback, setFeedback] = useState(null);

  const passwordPolicy = useMemo(
    () => 'Use at least 12 characters, mixing upper and lower case letters, numbers, and symbols.',
    []
  );

  const handleChange = (field) => (event) => {
    setFeedback(null);
    setForm((prev) => ({
      ...prev,
      [field]: event.target.value,
    }));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setFeedback(null);

    if (form.newPassword !== form.confirmPassword) {
      setFeedback({ type: 'error', message: 'New password and confirmation do not match.' });
      return;
    }

    setSubmitting(true);

    try {
      const response = await fetch('/api/auth/change-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          current_username: form.currentUsername,
          current_password: form.currentPassword,
          new_username: form.newUsername,
          new_password: form.newPassword,
          confirm_password: form.confirmPassword,
        }),
      });

      if (!response.ok) {
        const message = (await response.text()) || 'Failed to update credentials';
        throw new Error(message);
      }

      const data = await response.json();

      try {
        await login({ username: form.newUsername, password: form.newPassword });
      } catch (err) {
        setFeedback({
          type: 'warning',
          message: 'Credentials updated. Please sign in again with your new details.',
        });
        return;
      }

      setFeedback({
        type: 'success',
        message: data.message || 'Credentials updated successfully.',
        updatedAt: data.updated_at,
      });

      setForm((prev) => ({
        ...prev,
        currentUsername: prev.newUsername,
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
      }));
    } catch (error) {
      setFeedback({ type: 'error', message: error.message || 'Failed to update credentials.' });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Box>
      <Box display="flex" alignItems="center" justifyContent="space-between" mb={4}>
        <Box>
          <Typography variant="h4" gutterBottom>
            Settings
          </Typography>
          <Typography variant="body2" color="textSecondary">
            Manage dashboard preferences and harden access controls.
          </Typography>
        </Box>
      </Box>

      <Grid container spacing={4}>
        <Grid item xs={12} md={7}>
          <Paper
            component="form"
            onSubmit={handleSubmit}
            sx={{
              p: 4,
              borderRadius: 3,
              background: 'linear-gradient(160deg, rgba(99,102,241,0.16), rgba(15,23,42,0.92))',
              border: '1px solid rgba(99,102,241,0.25)',
              position: 'relative',
              overflow: 'hidden',
            }}
          >
            <Box display="flex" alignItems="center" gap={2} mb={3}>
              <Box
                sx={{
                  background: theme.palette.primary.main,
                  width: 44,
                  height: 44,
                  borderRadius: '16px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: '0 12px 24px rgba(99,102,241,0.35)',
                }}
              >
                <LockResetIcon fontSize="medium" />
              </Box>
              <Box>
                <Typography variant="h6" sx={{ fontWeight: 600 }}>
                  Account Security
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  Rotate dashboard credentials regularly to maintain least-privilege access.
                </Typography>
              </Box>
            </Box>

            {feedback && (
              <Alert severity={feedback.type} sx={{ mb: 3 }}>
                {feedback.message}
                {feedback.updatedAt && (
                  <Box component="span" ml={1} color={theme.palette.text.secondary}>
                    (Last updated {new Date(feedback.updatedAt).toLocaleString()})
                  </Box>
                )}
              </Alert>
            )}

            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <TextField
                  label="Current Username"
                  fullWidth
                  value={form.currentUsername}
                  onChange={handleChange('currentUsername')}
                  autoComplete="username"
                  required
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField
                  label="New Username"
                  fullWidth
                  value={form.newUsername}
                  onChange={handleChange('newUsername')}
                  autoComplete="username"
                  required
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField
                  label="Current Password"
                  type="password"
                  fullWidth
                  value={form.currentPassword}
                  onChange={handleChange('currentPassword')}
                  autoComplete="current-password"
                  required
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField
                  label="New Password"
                  type="password"
                  fullWidth
                  value={form.newPassword}
                  onChange={handleChange('newPassword')}
                  autoComplete="new-password"
                  helperText={passwordPolicy}
                  required
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  label="Confirm New Password"
                  type="password"
                  fullWidth
                  value={form.confirmPassword}
                  onChange={handleChange('confirmPassword')}
                  autoComplete="new-password"
                  required
                />
              </Grid>
            </Grid>

            <Divider sx={{ my: 3, borderColor: 'rgba(99,102,241,0.2)' }} />

            <Box display="flex" justifyContent="flex-end" gap={2}>
              <Button
                type="submit"
                variant="contained"
                disabled={submitting}
                sx={{ px: 4, py: 1 }}
              >
                {submitting ? 'Updating...' : 'Update Credentials'}
              </Button>
            </Box>
          </Paper>
        </Grid>

        <Grid item xs={12} md={5}>
          <Paper
            sx={{
              p: 4,
              borderRadius: 3,
              background: 'rgba(15,23,42,0.88)',
              border: '1px solid rgba(99,102,241,0.18)',
            }}
          >
            <Typography variant="h6" gutterBottom sx={{ fontWeight: 600 }}>
              Security Guidelines
            </Typography>
            <Typography variant="body2" color="textSecondary" paragraph>
              - Rotate credentials every quarter or whenever access is shared.
            </Typography>
            <Typography variant="body2" color="textSecondary" paragraph>
              - Avoid reusing credentials across environments. Prefer passphrases with
              12 or more characters.
            </Typography>
            <Typography variant="body2" color="textSecondary" paragraph>
              - Store updated credentials in your organization's password vault and
              restrict dashboard access to designated administrators.
            </Typography>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
}
