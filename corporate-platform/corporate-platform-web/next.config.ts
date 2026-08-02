import type { NextConfig } from "next";
import './src/env';

const nextConfig: NextConfig = {
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'images.unsplash.com',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: '*.pinata.cloud',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'cdn.jsdelivr.net',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: '*.stellar.org',
        pathname: '/**',
      },
    ],
    // Enable image optimization with caching
    unoptimized: false,
    minimumCacheTTL: 31536000, // 1 year for cached images
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
    formats: ['image/webp', 'image/avif'],
  },

  // Compression for text-based responses (gzip/brotli)
  compress: true,

  // Enable React Strict Mode for better development
  reactStrictMode: true,

  // Enable SWC minification for faster builds
  swcMinify: true,

  // Experimental features
  experimental: {
    // Optimize server components
    optimizeCss: true,
    // Enable Turbo for faster dev builds
    turbo: {
      resolveAlias: {
        '@': './src',
      },
    },
  },

  /**
   * Cache Headers Configuration
   * Defines caching strategies for different asset types to optimize CDN performance
   * and reduce load times.
   */
  async headers() {
    return [
      // Static assets (JS, CSS, chunks) - 1 year immutable
      {
        source: '/_next/static/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
          {
            key: 'Vary',
            value: 'Accept-Encoding',
          },
        ],
      },
      // Static assets (build ID) - 1 year immutable
      {
        source: '/_next/static/:buildId/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
          {
            key: 'Vary',
            value: 'Accept-Encoding',
          },
        ],
      },
      // Images from Next.js image optimization - 1 year immutable
      {
        source: '/_next/image/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
          {
            key: 'Vary',
            value: 'Accept-Encoding',
          },
        ],
      },
      // Fonts - 1 year immutable
      {
        source: '/_next/static/chunks/fonts/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
          {
            key: 'Vary',
            value: 'Accept-Encoding',
          },
        ],
      },
      {
        source: '/fonts/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
          {
            key: 'Vary',
            value: 'Accept-Encoding',
          },
        ],
      },
      // Favicon - 1 day
      {
        source: '/favicon.ico',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=86400, immutable',
          },
        ],
      },
      // Public assets - 1 day
      {
        source: '/public/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=86400',
          },
          {
            key: 'Vary',
            value: 'Accept-Encoding',
          },
        ],
      },
      // Manifest files - 1 hour
      {
        source: '/manifest.json',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=3600',
          },
        ],
      },
      // Site.webmanifest - 1 day
      {
        source: '/site.webmanifest',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=86400',
          },
        ],
      },
    ];
  },

  /**
   * Redirects Configuration
   */
  async redirects() {
    return [
      {
        source: '/corporate',
        destination: '/',
        permanent: true,
      },
      {
        source: '/dashboard',
        destination: '/',
        permanent: true,
      },
    ];
  },

  /**
   * Rewrites Configuration
   * Optional API proxy for development and production
   */
  async rewrites() {
    const apiUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:4000';
    return [
      {
        source: '/api/:path*',
        destination: `${apiUrl}/api/:path*`,
      },
    ];
  },

  // On-demand revalidation for ISR
  // This allows you to revalidate pages via API calls
  // https://nextjs.org/docs/app/building-your-application/data-fetching/revalidating#on-demand-revalidation
  // Uncomment and configure as needed:
  // experimental: {
  //   ...nextConfig.experimental,
  //   // Enable on-demand revalidation
  //   // revalidateSecret: process.env.REVALIDATE_SECRET,
  // },
};

export default nextConfig;