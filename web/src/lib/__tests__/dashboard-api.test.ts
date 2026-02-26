import { describe, it, expect } from 'vitest';
import { formatBytes } from '../dashboard-api';

describe('dashboard-api', () => {
  describe('formatBytes', () => {
    it('should return 0 B for zero', () => {
      expect(formatBytes(0)).toBe('0 B');
    });

    it('should format bytes', () => {
      expect(formatBytes(500)).toBe('500 B');
    });

    it('should format kilobytes', () => {
      expect(formatBytes(1024)).toBe('1 KB');
      expect(formatBytes(1536)).toBe('1.5 KB');
    });

    it('should format megabytes', () => {
      expect(formatBytes(1048576)).toBe('1 MB');
      expect(formatBytes(5242880)).toBe('5 MB');
    });

    it('should format gigabytes', () => {
      expect(formatBytes(1073741824)).toBe('1 GB');
      expect(formatBytes(2684354560)).toBe('2.5 GB');
    });

    it('should format terabytes', () => {
      expect(formatBytes(1099511627776)).toBe('1 TB');
    });
  });
});
