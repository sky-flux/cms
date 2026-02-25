import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Mock BlockNote modules
vi.mock('@blocknote/react', () => ({
  useCreateBlockNote: () => ({
    document: [],
    topLevelBlocks: [],
    domElement: { innerHTML: '<p>test content</p>' },
    replaceBlocks: vi.fn(),
    onChange: vi.fn(),
  }),
}));

vi.mock('@blocknote/shadcn', () => ({
  BlockNoteView: ({ children }: any) => (
    <div data-testid="blocknote-editor">{children}</div>
  ),
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      const map: Record<string, string> = {
        'content.postTitle': 'Title',
        'content.postTitlePlaceholder': 'Enter post title...',
        'content.postExcerpt': 'Excerpt',
        'content.postSlug': 'Slug',
        'content.postCoverImage': 'Cover Image',
        'content.postCategories': 'Categories',
        'content.postTags': 'Tags',
        'content.postSeo': 'SEO Settings',
        'content.postMetaTitle': 'Meta Title',
        'content.postMetaDescription': 'Meta Description',
        'content.postOgImage': 'OG Image URL',
        'content.publish': 'Publish',
        'content.versionConflict': 'This post was modified by another user. Please refresh.',
        'content.autoSaved': 'Auto-saved',
        'content.unsavedChanges': 'Unsaved changes',
        'common.save': 'Save',
        'common.create': 'Create',
        'messages.saveSuccess': 'Saved successfully.',
        'messages.createSuccess': 'Created successfully.',
      };
      return map[key] ?? key;
    },
  }),
}));

// Track mutations
const mockMutate = vi.fn();
const mockCreateMutate = vi.fn();
let mockQueryData: any = undefined;
let mockQueryLoading = false;

// Mock @tanstack/react-query
vi.mock('@tanstack/react-query', () => ({
  useQuery: ({ queryKey }: any) => ({
    data: mockQueryData,
    isLoading: mockQueryLoading,
    error: null,
  }),
  useMutation: ({ mutationFn, onSuccess, onError }: any) => {
    // Distinguish create vs update mutation based on context
    const isCreate = !mutationFn.toString().includes('update');
    const mutate = isCreate ? mockCreateMutate : mockMutate;
    return {
      mutate: (data: any) => {
        mutate(data);
        // Auto-invoke onSuccess for success path tests
        if ((mutate as any).__autoSuccess && onSuccess) {
          onSuccess({ data: { id: 'new-id', version: 1, title: 'Test' } });
        }
      },
      mutateAsync: async (data: any) => {
        mutate(data);
        return { data: { id: 'new-id', version: 1, title: 'Test' } };
      },
      isPending: false,
    };
  },
  useQueryClient: () => ({
    invalidateQueries: vi.fn(),
  }),
}));

// Mock postsApi
vi.mock('@/lib/content-api', () => ({
  postsApi: {
    get: vi.fn(),
    create: vi.fn().mockResolvedValue({ data: { id: 'new-id', version: 1, title: 'Test' } }),
    update: vi.fn().mockResolvedValue({ data: { id: 'post-1', version: 2, title: 'Updated' } }),
  },
  categoriesApi: {
    tree: vi.fn().mockResolvedValue({ data: [] }),
  },
  tagsApi: {
    suggest: vi.fn().mockResolvedValue({ data: [] }),
  },
}));

// Mock sonner
vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock editor store
vi.mock('@/stores/editor-store', () => ({
  useEditorStore: Object.assign(
    () => ({
      saveDraft: vi.fn(),
      getDraft: vi.fn(),
      clearDraft: vi.fn(),
    }),
    {
      getState: () => ({
        saveDraft: vi.fn(),
        getDraft: vi.fn(),
        clearDraft: vi.fn(),
      }),
    },
  ),
}));

import { PostEditor } from '../PostEditor';

describe('PostEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockQueryData = undefined;
    mockQueryLoading = false;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders title input', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByPlaceholderText('Enter post title...')).toBeInTheDocument();
  });

  it('renders BlockNote editor area', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByTestId('blocknote-editor')).toBeInTheDocument();
  });

  it('renders metadata panel with categories section', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByText('Categories')).toBeInTheDocument();
  });

  it('renders metadata panel with tags section', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByText('Tags')).toBeInTheDocument();
  });

  it('renders excerpt textarea in metadata panel', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByText('Excerpt')).toBeInTheDocument();
  });

  it('renders slug input in metadata panel', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByText('Slug')).toBeInTheDocument();
  });

  it('renders SEO settings section', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByText('SEO Settings')).toBeInTheDocument();
  });

  it('renders cover image input', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByText('Cover Image')).toBeInTheDocument();
  });

  it('shows Create button in create mode', () => {
    render(<PostEditor mode="create" />);
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument();
  });

  it('shows Save button in edit mode', () => {
    mockQueryData = {
      data: {
        id: 'post-1',
        title: 'Test Post',
        slug: 'test-post',
        content: '<p>Hello</p>',
        content_json: [{ type: 'paragraph', content: [{ type: 'text', text: 'Hello' }] }],
        excerpt: '',
        status: 'draft',
        version: 1,
        seo: null,
        categories: [],
        tags: [],
        cover_image: null,
      },
    };
    render(<PostEditor mode="edit" postId="post-1" />);
    expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
  });

  it('allows typing in title input', async () => {
    const user = userEvent.setup();
    render(<PostEditor mode="create" />);
    const titleInput = screen.getByPlaceholderText('Enter post title...');
    await user.type(titleInput, 'My New Post');
    expect(titleInput).toHaveValue('My New Post');
  });

  it('auto-generates slug from title in create mode', async () => {
    const user = userEvent.setup();
    render(<PostEditor mode="create" />);
    const titleInput = screen.getByPlaceholderText('Enter post title...');
    await user.type(titleInput, 'Hello World');
    // Slug should be auto-generated
    const slugInput = screen.getByTestId('slug-input');
    expect(slugInput).toHaveValue('hello-world');
  });

  it('populates fields when post data loads in edit mode', () => {
    mockQueryData = {
      data: {
        id: 'post-1',
        title: 'Existing Post',
        slug: 'existing-post',
        content: '<p>Content</p>',
        content_json: [],
        excerpt: 'An excerpt',
        status: 'draft',
        version: 3,
        seo: { meta_title: 'SEO Title', meta_description: 'SEO Desc', og_image_url: '' },
        categories: [{ id: 'cat-1', name: 'Tech', slug: 'tech' }],
        tags: [{ id: 'tag-1', name: 'React' }],
        cover_image: null,
      },
    };
    render(<PostEditor mode="edit" postId="post-1" />);
    expect(screen.getByPlaceholderText('Enter post title...')).toHaveValue('Existing Post');
  });

  it('renders two-column layout with editor and metadata panel', () => {
    const { container } = render(<PostEditor mode="create" />);
    const gridElement = container.querySelector('.grid');
    expect(gridElement).toBeInTheDocument();
  });

  // --- Auto-save and version conflict tests (Task 14) ---

  it('registers beforeunload handler', () => {
    const addEventSpy = vi.spyOn(window, 'addEventListener');
    render(<PostEditor mode="create" />);
    const beforeunloadCalls = addEventSpy.mock.calls.filter(
      ([event]) => event === 'beforeunload',
    );
    expect(beforeunloadCalls.length).toBeGreaterThanOrEqual(1);
    addEventSpy.mockRestore();
  });

  it('registers Ctrl+S / Cmd+S keyboard shortcut handler', () => {
    const addEventSpy = vi.spyOn(window, 'addEventListener');
    render(<PostEditor mode="create" />);
    const keydownCalls = addEventSpy.mock.calls.filter(
      ([event]) => event === 'keydown',
    );
    expect(keydownCalls.length).toBeGreaterThanOrEqual(1);
    addEventSpy.mockRestore();
  });

  it('calls save on Ctrl+S keypress in create mode', async () => {
    render(<PostEditor mode="create" />);
    // Dispatch Ctrl+S
    const event = new KeyboardEvent('keydown', {
      key: 's',
      ctrlKey: true,
      bubbles: true,
    });
    const preventDefaultSpy = vi.spyOn(event, 'preventDefault');
    window.dispatchEvent(event);
    expect(preventDefaultSpy).toHaveBeenCalled();
    // create mutation should have been called
    expect(mockCreateMutate).toHaveBeenCalled();
  });

  it('calls save on Cmd+S keypress in create mode', () => {
    render(<PostEditor mode="create" />);
    const event = new KeyboardEvent('keydown', {
      key: 's',
      metaKey: true,
      bubbles: true,
    });
    window.dispatchEvent(event);
    expect(mockCreateMutate).toHaveBeenCalled();
  });

  it('sets up auto-save interval for draft posts in edit mode', () => {
    const setIntervalSpy = vi.spyOn(globalThis, 'setInterval');
    mockQueryData = {
      data: {
        id: 'post-1',
        title: 'Draft Post',
        slug: 'draft-post',
        content: '',
        content_json: [],
        excerpt: '',
        status: 'draft',
        version: 1,
        seo: null,
        categories: [],
        tags: [],
        cover_image: null,
      },
    };
    render(<PostEditor mode="edit" postId="post-1" />);
    // Should have set up a 30-second interval for auto-save
    const intervalCalls = setIntervalSpy.mock.calls.filter(
      ([, interval]) => interval === 30000,
    );
    expect(intervalCalls.length).toBeGreaterThanOrEqual(1);
    setIntervalSpy.mockRestore();
  });

  it('does not set up auto-save interval for non-draft posts', () => {
    const setIntervalSpy = vi.spyOn(globalThis, 'setInterval');
    mockQueryData = {
      data: {
        id: 'post-1',
        title: 'Published Post',
        slug: 'published-post',
        content: '',
        content_json: [],
        excerpt: '',
        status: 'published',
        version: 1,
        seo: null,
        categories: [],
        tags: [],
        cover_image: null,
      },
    };
    render(<PostEditor mode="edit" postId="post-1" />);
    const intervalCalls = setIntervalSpy.mock.calls.filter(
      ([, interval]) => interval === 30000,
    );
    expect(intervalCalls.length).toBe(0);
    setIntervalSpy.mockRestore();
  });

  it('cleans up auto-save interval on unmount', () => {
    const clearIntervalSpy = vi.spyOn(globalThis, 'clearInterval');
    mockQueryData = {
      data: {
        id: 'post-1',
        title: 'Draft Post',
        slug: 'draft-post',
        content: '',
        content_json: [],
        excerpt: '',
        status: 'draft',
        version: 1,
        seo: null,
        categories: [],
        tags: [],
        cover_image: null,
      },
    };
    const { unmount } = render(<PostEditor mode="edit" postId="post-1" />);
    unmount();
    expect(clearIntervalSpy).toHaveBeenCalled();
    clearIntervalSpy.mockRestore();
  });

  it('calls update mutation with version when Save button clicked in edit mode', async () => {
    const user = userEvent.setup();
    mockQueryData = {
      data: {
        id: 'post-1',
        title: 'Test Post',
        slug: 'test-post',
        content: '',
        content_json: [],
        excerpt: '',
        status: 'draft',
        version: 5,
        seo: null,
        categories: [],
        tags: [],
        cover_image: null,
      },
    };
    render(<PostEditor mode="edit" postId="post-1" />);
    await user.click(screen.getByRole('button', { name: /save/i }));
    expect(mockMutate).toHaveBeenCalledWith(
      expect.objectContaining({ version: 5 }),
    );
  });

  it('stops auto-slug generation when slug is manually edited', async () => {
    const user = userEvent.setup();
    render(<PostEditor mode="create" />);
    const slugInput = screen.getByTestId('slug-input');
    // First manually edit the slug
    await user.clear(slugInput);
    await user.type(slugInput, 'custom-slug');
    // Then change the title — slug should NOT change
    const titleInput = screen.getByPlaceholderText('Enter post title...');
    await user.type(titleInput, 'New Title');
    expect(slugInput).toHaveValue('custom-slug');
  });
});
