// Shared jobDefinitions for jobs.html, activity.html, activity_table.html
window.jobDefinitions = {
  'trending_movies': {
    name: 'Trending Movies',
    description: 'Movies currently being watched and talked about on Trakt',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path></svg>',
    iconColor: 'text-yellow-400',
    gradient: 'from-yellow-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.trending_movies'
  },
  'popular_movies': {
    name: 'Popular Movies',
    description: 'Most watched movies on Trakt over various time periods',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.879 16.121A3 3 0 1012.015 11L11 14H9c0 .768.293 1.536.879 2.121z"></path></svg>',
    iconColor: 'text-red-400',
    gradient: 'from-red-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.popular_movies'
  },
  'box_office': {
    name: 'Box Office',
    description: 'Top 10 movies from the weekend box office',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z"></path></svg>',
    iconColor: 'text-green-400',
    gradient: 'from-green-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.box_office'
  },
  'anticipated_movies': {
    name: 'Anticipated Movies',
    description: 'Most anticipated upcoming movies',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path></svg>',
    iconColor: 'text-purple-400',
    gradient: 'from-purple-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.anticipated_movies'
  },
  'favorited_movies': {
    name: 'Favorited Movies',
    description: 'Most favorited movies on Trakt',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path></svg>',
    iconColor: 'text-pink-400',
    gradient: 'from-pink-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.favorited_movies'
  },
  'played_movies': {
    name: 'Played Movies',
    description: 'Most played movies over a time period',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>',
    iconColor: 'text-blue-400',
    gradient: 'from-blue-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.played_movies'
  },
  'watched_movies': {
    name: 'Watched Movies',
    description: 'Most watched movies over a time period',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path></svg>',
    iconColor: 'text-cyan-400',
    gradient: 'from-cyan-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.watched_movies'
  },
  'collected_movies': {
    name: 'Collected Movies',
    description: 'Most collected movies over a time period',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"></path></svg>',
    iconColor: 'text-amber-400',
    gradient: 'from-amber-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.collected_movies'
  },
  'smart_popular_movies': {
    name: 'Smart Popular Movies',
    description: 'Adaptive popular movies with dynamic rating thresholds',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"></path></svg>',
    iconColor: 'text-indigo-400',
    gradient: 'from-indigo-900/50 to-slate-800/50',
    category: 'Movies',
    configKey: 'jobs.smart_popular_movies'
  },
  'trending_shows': {
    name: 'Trending Shows',
    description: 'TV shows currently being watched and talked about on Trakt',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path></svg>',
    iconColor: 'text-yellow-400',
    gradient: 'from-yellow-900/50 to-slate-800/50',
    category: 'TV Shows',
    configKey: 'jobs.trending_shows'
  },
  'popular_shows': {
    name: 'Popular Shows',
    description: 'Most watched TV shows on Trakt over various time periods',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.879 16.121A3 3 0 1012.015 11L11 14H9c0 .768.293 1.536.879 2.121z"></path></svg>',
    iconColor: 'text-red-400',
    gradient: 'from-red-900/50 to-slate-800/50',
    category: 'TV Shows',
    configKey: 'jobs.popular_shows'
  },
  'anticipated_shows': {
    name: 'Anticipated Shows',
    description: 'Most anticipated upcoming TV shows',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path></svg>',
    iconColor: 'text-purple-400',
    gradient: 'from-purple-900/50 to-slate-800/50',
    category: 'TV Shows',
    configKey: 'jobs.anticipated_shows'
  },
  'favorited_shows': {
    name: 'Favorited Shows',
    description: 'Most favorited TV shows on Trakt',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path></svg>',
    iconColor: 'text-pink-400',
    gradient: 'from-pink-900/50 to-slate-800/50',
    category: 'TV Shows',
    configKey: 'jobs.favorited_shows'
  },
  'played_shows': {
    name: 'Played Shows',
    description: 'Most played TV shows over a time period',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>',
    iconColor: 'text-blue-400',
    gradient: 'from-blue-900/50 to-slate-800/50',
    category: 'TV Shows',
    configKey: 'jobs.played_shows'
  },
  'watched_shows': {
    name: 'Watched Shows',
    description: 'Most watched TV shows over a time period',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path></svg>',
    iconColor: 'text-cyan-400',
    gradient: 'from-cyan-900/50 to-slate-800/50',
    category: 'TV Shows',
    configKey: 'jobs.watched_shows'
  },
  'collected_shows': {
    name: 'Collected Shows',
    description: 'Most collected TV shows over a time period',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"></path></svg>',
    iconColor: 'text-amber-400',
    gradient: 'from-amber-900/50 to-slate-800/50',
    category: 'TV Shows',
    configKey: 'jobs.collected_shows'
  },
  'smart_popular_shows': {
    name: 'Smart Popular Shows',
    description: 'Adaptive popular TV shows with dynamic rating thresholds',
    icon: '<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"></path></svg>',
    iconColor: 'text-indigo-400',
    gradient: 'from-indigo-900/50 to-slate-800/50',
    category: 'TV Shows',
    configKey: 'jobs.smart_popular_shows'
  }
};
