import { z } from 'zod';

export const eventSchema = z.object({
	name: z.string().min(1, 'Event name is required').max(255),
	slug: z.string().min(1, 'Slug is required').max(255).toLowerCase().regex(/^[a-z0-9-]+$/, 'Slug must be lowercase with hyphens only'),
	description: z.string().optional(),
	location: z.string().optional(),
	date: z.string().optional(),
	sport: z.string().min(1, 'Sport is required').max(255),
	status: z.enum(['draft', 'published', 'archived']).default('draft'),
	tags: z.array(z.string()).default([]),
});

export const gallerySchema = z.object({
	name: z.string().min(1, 'Gallery name is required').max(255),
	description: z.string().optional(),
});

export type EventSchema = typeof eventSchema;
export type GallerySchema = typeof gallerySchema;
