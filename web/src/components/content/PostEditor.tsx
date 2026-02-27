import { useState, useEffect, useCallback, useRef, lazy, Suspense } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCreateBlockNote } from '@blocknote/react';
import { toast } from 'sonner';
import { postsApi } from '@/lib/content-api';
import type { UpdatePostDTO, CreatePostDTO } from '@/lib/content-api';
import { ApiError } from '@/lib/api-client';
import { useEditorStore } from '@/stores/editor-store';
import { CategorySelect } from './CategorySelect';
import { TagSelect } from './TagSelect';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Card } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';

// Lazy load BlockNote editor view (1MB+)
const BlockNoteView = lazy(() => import('@blocknote/shadcn').then(m => ({ default: m.BlockNoteView })));

// Load BlockNote CSS globally
import '@blocknote/shadcn/style.css';

function EditorLoading() {
  const { t } = useTranslation();
  return (
    <div className="flex items-center justify-center h-[400px] border rounded-lg bg-muted/20">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2" />
        <p className="text-sm text-muted-foreground">{t('common.loading')}</p>
      </div>
    </div>
  );
}

interface PostEditorProps {
  mode: 'create' | 'edit';
  postId?: string;
  onCreated?: (postId: string) => void;
}

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

export function PostEditor({ mode, postId, onCreated }: PostEditorProps) {
  const { t } = useTranslation();
  const _queryClient = useQueryClient();
  const editorStore = useEditorStore();

  // Form state
  const [title, setTitle] = useState('');
  const [slug, setSlug] = useState('');
  const [excerpt, setExcerpt] = useState('');
  const [coverImageUrl, setCoverImageUrl] = useState('');
  const [categoryIds, setCategoryIds] = useState<string[]>([]);
  const [tagIds, setTagIds] = useState<string[]>([]);
  const [metaTitle, setMetaTitle] = useState('');
  const [metaDescription, setMetaDescription] = useState('');
  const [ogImageUrl, setOgImageUrl] = useState('');
  const [seoOpen, setSeoOpen] = useState(false);
  const [version, setVersion] = useState(1);
  const [status, setStatus] = useState('draft');
  const [isDirty, setIsDirty] = useState(false);
  const [autoSlug, setAutoSlug] = useState(mode === 'create');
  const lastSavedRef = useRef<string>('');
  const autoSaveTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Initial content state for BlockNote editor
  const [initialContent, setInitialContent] = useState<any[] | undefined | "loading">(
    mode === 'edit' ? "loading" : undefined
  );

  // Fetch post data in edit mode
  const { data: postData, isLoading: postLoading } = useQuery({
    queryKey: ['post', postId],
    queryFn: () => postsApi.get(postId!),
    enabled: mode === 'edit' && !!postId,
  });

  // Load initial content when post data is fetched
  useEffect(() => {
    if (mode === 'edit' && postData?.data) {
      const post = postData.data;
      console.log('[PostEditor] Setting initialContent:', post.content_json);
      if (post.content_json && Array.isArray(post.content_json)) {
        setInitialContent(post.content_json);
      } else {
        setInitialContent(undefined);
      }
    }
  }, [mode, postData]);

  // BlockNote editor - only create when initialContent is ready
  const editor = useCreateBlockNote({
    initialContent: initialContent === "loading" ? undefined : initialContent,
  });

  // Update editor content when initialContent changes (after editor is created)
  useEffect(() => {
    if (editor && initialContent && initialContent !== "loading" && Array.isArray(initialContent)) {
      editor.replaceBlocks(editor.document, initialContent);
    }
  }, [editor, initialContent]);

  // Populate form fields from loaded post
  useEffect(() => {
    if (postData?.data) {
      const post = postData.data;
      setTitle(post.title);
      setSlug(post.slug);
      setExcerpt(post.excerpt || '');
      setVersion(post.version);
      setStatus(post.status);
      setCategoryIds(post.categories?.map((c: any) => c.id) || []);
      setTagIds(post.tags?.map((t: any) => t.id) || []);
      if (post.seo) {
        setMetaTitle(post.seo.meta_title || '');
        setMetaDescription(post.seo.meta_description || '');
        setOgImageUrl(post.seo.og_image_url || '');
      }
      if (post.cover_image) {
        setCoverImageUrl(post.cover_image.url || '');
      }
      setAutoSlug(false);
      lastSavedRef.current = JSON.stringify(post.content_json);
    }
  }, [postData]);

  // Auto-generate slug from title in create mode
  useEffect(() => {
    if (autoSlug) {
      setSlug(slugify(title));
    }
  }, [title, autoSlug]);

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: CreatePostDTO) => postsApi.create(data),
    onSuccess: (result) => {
      toast.success(t('messages.createSuccess'));
      setIsDirty(false);
      if (onCreated && result?.data?.id) {
        onCreated(result.data.id);
      }
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: (data: UpdatePostDTO) => postsApi.update(postId!, data),
    onSuccess: (result) => {
      toast.success(t('messages.saveSuccess'));
      setIsDirty(false);
      if (result?.data?.version) {
        setVersion(result.data.version);
      }
      lastSavedRef.current = JSON.stringify(editor.document);
    },
    onError: (error: Error) => {
      if (error instanceof ApiError && error.status === 409) {
        toast.error(t('content.versionConflict'));
      } else {
        toast.error(error.message);
      }
    },
  });

  // Build save data
  const buildPostData = useCallback(() => {
    const contentJson = editor.document;
    const contentHtml = editor.domElement?.innerHTML || '';
    console.log('[PostEditor] Building save data, editor.document:', contentJson);
    console.log('[PostEditor] Editor domElement:', editor.domElement);
    return {
      title,
      slug,
      content: contentHtml,
      content_json: contentJson,
      excerpt,
      category_ids: categoryIds.length > 0 ? categoryIds : undefined,
      tag_ids: tagIds.length > 0 ? tagIds : undefined,
      meta_title: metaTitle || undefined,
      meta_description: metaDescription || undefined,
      og_image_url: ogImageUrl || undefined,
    };
  }, [title, slug, excerpt, categoryIds, tagIds, metaTitle, metaDescription, ogImageUrl, editor]);

  // Handle save
  const handleSave = useCallback(() => {
    if (mode === 'create') {
      createMutation.mutate(buildPostData());
    } else {
      updateMutation.mutate({ ...buildPostData(), version });
    }
  }, [mode, buildPostData, version, createMutation, updateMutation]);

  // Auto-save for drafts (30s interval) — use post data status directly
  const postStatus = postData?.data?.status;
  useEffect(() => {
    if (mode === 'edit' && postStatus === 'draft' && postId) {
      autoSaveTimerRef.current = setInterval(() => {
        const currentContent = JSON.stringify(editor.document);
        if (currentContent !== lastSavedRef.current) {
          editorStore.saveDraft(postId, currentContent);
          updateMutation.mutate({ ...buildPostData(), version });
        }
      }, 30000);

      return () => {
        if (autoSaveTimerRef.current) {
          clearInterval(autoSaveTimerRef.current);
        }
      };
    }
  }, [mode, postStatus, postId, editor, version, buildPostData, editorStore, updateMutation]);

  // Dirty tracking
  useEffect(() => {
    const currentContent = JSON.stringify(editor.document);
    setIsDirty(currentContent !== lastSavedRef.current || title !== (postData?.data?.title || ''));
  }, [title, editor, postData]);

  // beforeunload warning on unsaved changes
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (isDirty) {
        e.preventDefault();
      }
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [isDirty]);

  // Keyboard shortcut: Ctrl+S / Cmd+S
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        handleSave();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [handleSave]);

  return (
    <div className="grid grid-cols-[1fr_320px] gap-6 h-full">
      {/* Left: Editor */}
      <div className="flex flex-col gap-4 min-w-0">
        <Input
          value={title}
          onChange={(e) => {
            setTitle(e.target.value);
            setIsDirty(true);
          }}
          placeholder={t('content.postTitlePlaceholder')}
          className="text-2xl font-bold h-14 border-0 border-b rounded-none focus-visible:ring-0 px-0"
        />
        <div className="flex-1 min-h-[400px] border rounded-lg overflow-hidden">
          {initialContent === "loading" ? (
            <EditorLoading />
          ) : (
            <Suspense fallback={<EditorLoading />}>
              <BlockNoteView
                editor={editor}
                theme="light"
              />
            </Suspense>
          )}
        </div>
      </div>

      {/* Right: Metadata Panel */}
      <div className="flex flex-col gap-4 overflow-y-auto" data-testid="metadata-panel">
        {/* Save/Create Button */}
        <Button
          onClick={handleSave}
          disabled={createMutation.isPending || updateMutation.isPending}
          className="w-full"
        >
          {mode === 'create' ? t('common.create') : t('common.save')}
        </Button>

        <Separator />

        {/* Categories */}
        <Card className="p-4">
          <Label className="text-sm font-medium">{t('content.postCategories')}</Label>
          <div className="mt-2">
            <CategorySelect value={categoryIds} onChange={setCategoryIds} />
          </div>
        </Card>

        {/* Tags */}
        <Card className="p-4">
          <Label className="text-sm font-medium">{t('content.postTags')}</Label>
          <div className="mt-2">
            <TagSelect value={tagIds} onChange={setTagIds} />
          </div>
        </Card>

        {/* Cover Image */}
        <Card className="p-4">
          <Label className="text-sm font-medium">{t('content.postCoverImage')}</Label>
          <Input
            placeholder="https://..."
            className="mt-2"
            value={coverImageUrl}
            onChange={(e) => setCoverImageUrl(e.target.value)}
          />
        </Card>

        {/* Excerpt */}
        <Card className="p-4">
          <Label className="text-sm font-medium">{t('content.postExcerpt')}</Label>
          <textarea
            className="mt-2 w-full min-h-[80px] rounded-md border border-input bg-background px-3 py-2 text-sm"
            value={excerpt}
            onChange={(e) => setExcerpt(e.target.value)}
          />
        </Card>

        {/* Slug */}
        <Card className="p-4">
          <Label className="text-sm font-medium">{t('content.postSlug')}</Label>
          <Input
            data-testid="slug-input"
            className="mt-2"
            value={slug}
            onChange={(e) => {
              setSlug(e.target.value);
              setAutoSlug(false);
            }}
          />
        </Card>

        {/* SEO Settings (Collapsible) */}
        <Card className="p-4">
          <button
            type="button"
            onClick={() => setSeoOpen(!seoOpen)}
            className="w-full text-left text-sm font-medium flex items-center justify-between"
          >
            {t('content.postSeo')}
            <span className="text-muted-foreground">{seoOpen ? '-' : '+'}</span>
          </button>
          {seoOpen && (
            <div className="mt-3 flex flex-col gap-3">
              <div>
                <Label className="text-xs">{t('content.postMetaTitle')}</Label>
                <Input
                  className="mt-1"
                  value={metaTitle}
                  onChange={(e) => setMetaTitle(e.target.value)}
                />
              </div>
              <div>
                <Label className="text-xs">{t('content.postMetaDescription')}</Label>
                <textarea
                  className="mt-1 w-full min-h-[60px] rounded-md border border-input bg-background px-3 py-2 text-xs"
                  value={metaDescription}
                  onChange={(e) => setMetaDescription(e.target.value)}
                />
              </div>
              <div>
                <Label className="text-xs">{t('content.postOgImage')}</Label>
                <Input
                  className="mt-1"
                  value={ogImageUrl}
                  onChange={(e) => setOgImageUrl(e.target.value)}
                />
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
