// import { type PageServerLoad } from './$types'




// export const load = async ({ params, fetch, parent }) => {
//   const { event } = await parent()
//   const res = await fetch(`/api/v1/events/${event.id}/galleries/${params.galleryId}/photos`)
//   const { data: photos, next_cursor } = await res.json()
//   return { event, photos, next_cursor }
// }