import { useState, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from 'react-i18next';
import { Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { api, ApiError } from '@/lib/api-client';

// --- Schemas ---

const step1Schema = z.object({
  admin_display_name: z.string().min(1).max(100),
  admin_email: z.string().email(),
  password: z.string().min(8),
  confirmPassword: z.string().min(8),
}).refine((d) => d.password === d.confirmPassword, {
  path: ['confirmPassword'],
  message: 'Passwords do not match',
});

const step2Schema = z.object({
  site_name: z.string().min(1).max(200),
  site_slug: z.string().regex(/^[a-z0-9-]{3,50}$/),
  site_url: z.string().url(),
  locale: z.string().optional(),
});

type Step1Values = z.infer<typeof step1Schema>;
type Step2Values = z.infer<typeof step2Schema>;

interface FormData {
  admin_display_name: string;
  admin_email: string;
  admin_password: string;
  site_name: string;
  site_slug: string;
  site_url: string;
  locale: string;
}

// --- Step Components ---

function Step1Form({
  defaultValues,
  onNext,
  t,
}: {
  defaultValues: { admin_display_name: string; admin_email: string; password: string; confirmPassword: string };
  onNext: (values: Step1Values) => void;
  t: (key: string) => string;
}) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Step1Values>({
    resolver: zodResolver(step1Schema),
    defaultValues,
  });

  return (
    <form onSubmit={handleSubmit(onNext)} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="admin_display_name">{t('auth.setupAdminUsername')}</Label>
        <Input
          id="admin_display_name"
          placeholder={t('auth.setupAdminUsernamePlaceholder')}
          {...register('admin_display_name')}
          aria-invalid={!!errors.admin_display_name}
        />
        {errors.admin_display_name && (
          <p className="text-sm text-destructive">{errors.admin_display_name.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="admin_email">{t('auth.setupAdminEmail')}</Label>
        <Input
          id="admin_email"
          type="email"
          placeholder={t('auth.setupAdminEmailPlaceholder')}
          {...register('admin_email')}
          aria-invalid={!!errors.admin_email}
        />
        {errors.admin_email && (
          <p className="text-sm text-destructive">{errors.admin_email.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">{t('auth.password')}</Label>
        <Input
          id="password"
          type="password"
          placeholder={t('auth.passwordPlaceholder')}
          {...register('password')}
          aria-invalid={!!errors.password}
        />
        {errors.password && (
          <p className="text-sm text-destructive">{t('auth.passwordMinLength')}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="confirmPassword">{t('auth.confirmPassword')}</Label>
        <Input
          id="confirmPassword"
          type="password"
          {...register('confirmPassword')}
          aria-invalid={!!errors.confirmPassword}
        />
        {errors.confirmPassword && (
          <p className="text-sm text-destructive">{t('auth.passwordsDoNotMatch')}</p>
        )}
      </div>

      <div className="flex justify-end">
        <Button type="submit">{t('auth.setupNext')}</Button>
      </div>
    </form>
  );
}

function Step2Form({
  defaultValues,
  onNext,
  onBack,
  t,
}: {
  defaultValues: { site_name: string; site_slug: string; site_url: string; locale: string };
  onNext: (values: Step2Values) => void;
  onBack: () => void;
  t: (key: string) => string;
}) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Step2Values>({
    resolver: zodResolver(step2Schema),
    defaultValues,
  });

  return (
    <form onSubmit={handleSubmit(onNext)} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="site_name">{t('auth.setupSiteName')}</Label>
        <Input
          id="site_name"
          placeholder={t('auth.setupSiteNamePlaceholder')}
          {...register('site_name')}
          aria-invalid={!!errors.site_name}
        />
        {errors.site_name && (
          <p className="text-sm text-destructive">{errors.site_name.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="site_slug">{t('auth.setupSiteSlug')}</Label>
        <Input
          id="site_slug"
          placeholder={t('auth.setupSiteSlugPlaceholder')}
          {...register('site_slug')}
          aria-invalid={!!errors.site_slug}
        />
        {errors.site_slug && (
          <p className="text-sm text-destructive">{t('auth.slugFormat')}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="site_url">{t('auth.setupSiteUrl')}</Label>
        <Input
          id="site_url"
          placeholder={t('auth.setupSiteUrlPlaceholder')}
          {...register('site_url')}
          aria-invalid={!!errors.site_url}
        />
        {errors.site_url && (
          <p className="text-sm text-destructive">{errors.site_url.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="locale">{t('auth.setupLocale')}</Label>
        <select
          id="locale"
          className="h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs"
          {...register('locale')}
        >
          <option value="en">English</option>
          <option value="zh-CN">中文</option>
        </select>
      </div>

      <div className="flex justify-between">
        <Button type="button" variant="outline" onClick={onBack}>
          {t('auth.setupBack')}
        </Button>
        <Button type="submit">{t('auth.setupNext')}</Button>
      </div>
    </form>
  );
}

function Step3Review({
  data,
  onBack,
  onInstall,
  isInstalling,
  t,
}: {
  data: FormData;
  onBack: () => void;
  onInstall: () => void;
  isInstalling: boolean;
  t: (key: string) => string;
}) {
  return (
    <div className="space-y-4">
      <div className="space-y-3 rounded-md border p-4 text-sm">
        <div className="flex justify-between">
          <span className="text-muted-foreground">{t('auth.setupAdminUsername')}</span>
          <span>{data.admin_display_name}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">{t('auth.setupAdminEmail')}</span>
          <span>{data.admin_email}</span>
        </div>
        <Separator />
        <div className="flex justify-between">
          <span className="text-muted-foreground">{t('auth.setupSiteName')}</span>
          <span>{data.site_name}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">{t('auth.setupSiteSlug')}</span>
          <span>{data.site_slug}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">{t('auth.setupSiteUrl')}</span>
          <span>{data.site_url}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-muted-foreground">{t('auth.setupLocale')}</span>
          <span>{data.locale}</span>
        </div>
      </div>

      <div className="flex justify-between">
        <Button type="button" variant="outline" onClick={onBack}>
          {t('auth.setupBack')}
        </Button>
        <Button onClick={onInstall} disabled={isInstalling}>
          {isInstalling && <Loader2 className="mr-2 size-4 animate-spin" />}
          {isInstalling ? t('auth.setupInstalling') : t('auth.setupInstall')}
        </Button>
      </div>
    </div>
  );
}

// --- Main Component ---

export function SetupWizard() {
  const { t } = useTranslation();
  const [step, setStep] = useState(1);
  const [isInstalling, setIsInstalling] = useState(false);

  const formData = useRef<FormData>({
    admin_display_name: '',
    admin_email: '',
    admin_password: '',
    site_name: '',
    site_slug: '',
    site_url: '',
    locale: 'en',
  });

  const handleStep1Next = (values: Step1Values) => {
    formData.current.admin_display_name = values.admin_display_name;
    formData.current.admin_email = values.admin_email;
    formData.current.admin_password = values.password;
    setStep(2);
  };

  const handleStep2Next = (values: Step2Values) => {
    formData.current.site_name = values.site_name;
    formData.current.site_slug = values.site_slug;
    formData.current.site_url = values.site_url;
    formData.current.locale = values.locale || 'en';
    setStep(3);
  };

  const handleInstall = async () => {
    setIsInstalling(true);
    try {
      await api.post('/api/v1/setup/initialize', {
        admin_display_name: formData.current.admin_display_name,
        admin_email: formData.current.admin_email,
        admin_password: formData.current.admin_password,
        site_name: formData.current.site_name,
        site_slug: formData.current.site_slug,
        site_url: formData.current.site_url,
        locale: formData.current.locale,
      });
      window.location.href = '/setup/complete';
    } catch (err) {
      setIsInstalling(false);
      if (err instanceof ApiError) {
        toast.error(err.message);
      } else {
        toast.error(t('errors.networkError'));
      }
    }
  };

  const stepLabels = [t('auth.setupStep1'), t('auth.setupStep2'), t('auth.setupStep3')];

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle>{t('auth.setupTitle')}</CardTitle>
        <CardDescription>{t('auth.setupDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        {/* Step indicator */}
        <div className="mb-6">
          <div className="flex items-center justify-center gap-2 text-sm">
            {stepLabels.map((_, i) => (
              <span
                key={i}
                className={`flex size-6 items-center justify-center rounded-full text-xs ${
                  i + 1 <= step
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted text-muted-foreground'
                }`}
              >
                {i + 1}
              </span>
            ))}
          </div>
          <Separator className="mt-4" />
        </div>

        {/* Current step label */}
        <h3 className="mb-4 text-center font-medium">{stepLabels[step - 1]}</h3>

        {/* Step forms */}
        {step === 1 && (
          <Step1Form
            defaultValues={{
              admin_display_name: formData.current.admin_display_name,
              admin_email: formData.current.admin_email,
              password: formData.current.admin_password,
              confirmPassword: formData.current.admin_password,
            }}
            onNext={handleStep1Next}
            t={t}
          />
        )}

        {step === 2 && (
          <Step2Form
            defaultValues={{
              site_name: formData.current.site_name,
              site_slug: formData.current.site_slug,
              site_url: formData.current.site_url,
              locale: formData.current.locale,
            }}
            onNext={handleStep2Next}
            onBack={() => setStep(1)}
            t={t}
          />
        )}

        {step === 3 && (
          <Step3Review
            data={formData.current}
            onBack={() => setStep(2)}
            onInstall={handleInstall}
            isInstalling={isInstalling}
            t={t}
          />
        )}
      </CardContent>
    </Card>
  );
}
