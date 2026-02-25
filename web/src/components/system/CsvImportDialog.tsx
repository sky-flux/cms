import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Upload } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';

interface CsvImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onImport: (file: File) => void | Promise<void>;
  loading?: boolean;
}

function parseCSVPreview(text: string, maxRows = 10): string[][] {
  const lines = text.split('\n').filter(Boolean);
  return lines.slice(0, maxRows + 1).map((line) => line.split(',').map((s) => s.trim()));
}

export function CsvImportDialog({
  open,
  onOpenChange,
  onImport,
  loading = false,
}: CsvImportDialogProps) {
  const { t } = useTranslation();
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string[][] | null>(null);

  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0] ?? null;
    setFile(selectedFile);

    if (selectedFile) {
      const reader = new FileReader();
      reader.onload = (evt) => {
        const text = evt.target?.result as string;
        if (text) {
          setPreview(parseCSVPreview(text));
        }
      };
      reader.readAsText(selectedFile);
    } else {
      setPreview(null);
    }
  }, []);

  const handleImport = useCallback(async () => {
    if (file) {
      await onImport(file);
    }
  }, [file, onImport]);

  const handleOpenChange = useCallback(
    (isOpen: boolean) => {
      if (!isOpen) {
        setFile(null);
        setPreview(null);
      }
      onOpenChange(isOpen);
    },
    [onOpenChange],
  );

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('system.redirects.importCsv')}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="csv-file">{t('system.redirects.csvFile')}</Label>
            <input
              id="csv-file"
              type="file"
              accept=".csv"
              aria-label={t('system.redirects.csvFile')}
              onChange={handleFileChange}
              className="block w-full text-sm text-muted-foreground file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-medium file:bg-primary file:text-primary-foreground hover:file:bg-primary/90"
            />
            <p className="text-xs text-muted-foreground">
              {t('system.redirects.csvFormat')}
            </p>
          </div>

          {/* CSV Preview */}
          {preview && preview.length > 0 && (
            <div className="space-y-2">
              <Label>{t('system.redirects.csvPreview')}</Label>
              <div className="max-h-48 overflow-auto rounded-md border">
                <table className="w-full text-xs">
                  <thead className="bg-muted">
                    <tr>
                      {preview[0].map((header, i) => (
                        <th key={i} className="px-2 py-1 text-left font-medium">
                          {header}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {preview.slice(1).map((row, i) => (
                      <tr key={i} className="border-t">
                        {row.map((cell, j) => (
                          <td key={j} className="px-2 py-1">
                            {cell}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleImport}
            disabled={!file || loading}
            aria-label={t('system.redirects.importCsv')}
          >
            {loading ? (
              t('common.loading')
            ) : (
              <>
                <Upload className="mr-2 h-4 w-4" />
                {t('system.redirects.importCsv')}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
