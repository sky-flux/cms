import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Loader2 } from 'lucide-react';
import { toast, Toaster } from 'sonner';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from '@/components/ui/input-otp';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { api, ApiError } from '@/lib/api-client';
import { useAuthStore } from '@/stores/auth-store';
import type { LoginSuccessData } from '@/lib/auth-api';

interface TwoFactorFormProps {
  tempToken: string;
}

function TwoFactorFormInner({ tempToken }: TwoFactorFormProps) {
  const { t } = useTranslation();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const setAuth = useAuthStore((s) => s.setAuth);

  const handleComplete = async (code: string) => {
    if (code.length !== 6) return;
    setIsSubmitting(true);

    try {
      const resp = await api.post<{
        success: boolean;
        data: LoginSuccessData;
      }>('/v1/auth/2fa/validate', { code }, {
        headers: { Authorization: `Bearer ${tempToken}` },
      });

      const data = resp.data;
      setAuth(
        {
          id: data.user.id,
          email: data.user.email,
          displayName: data.user.display_name,
          avatarUrl: '',
        },
        data.access_token,
      );
      window.location.href = '/dashboard';
    } catch (err) {
      setIsSubmitting(false);
      if (err instanceof ApiError) {
        toast.error(err.message);
      } else {
        toast.error(t('errors.networkError'));
      }
    }
  };

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle>{t('auth.twoFactorTitle')}</CardTitle>
        <CardDescription>{t('auth.twoFactorDescription')}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        {isSubmitting ? (
          <Loader2 className="size-8 animate-spin text-primary" />
        ) : (
          <InputOTP maxLength={6} onComplete={handleComplete}>
            <InputOTPGroup>
              <InputOTPSlot index={0} />
              <InputOTPSlot index={1} />
              <InputOTPSlot index={2} />
              <InputOTPSlot index={3} />
              <InputOTPSlot index={4} />
              <InputOTPSlot index={5} />
            </InputOTPGroup>
          </InputOTP>
        )}

        <a
          href="/login"
          className="text-sm text-muted-foreground hover:text-primary"
        >
          {t('auth.twoFactorBackToLogin')}
        </a>
      </CardContent>
    </Card>
  );
}

export function TwoFactorForm({ tempToken }: TwoFactorFormProps) {
  return (
    <I18nProvider>
      <TwoFactorFormInner tempToken={tempToken} />
      <Toaster />
    </I18nProvider>
  );
}
