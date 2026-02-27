import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { LoginForm } from '../LoginForm';
import { QueryProvider } from '@/features/shared/components/QueryProvider';

describe('LoginForm', () => {
  it('renders email and password fields', () => {
    render(
      <QueryProvider>
        <LoginForm />
      </QueryProvider>
    );
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it('shows submit button', () => {
    render(
      <QueryProvider>
        <LoginForm />
      </QueryProvider>
    );
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });
});
