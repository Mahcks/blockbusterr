// Shared jobDefinitions for jobs.html, activity.html, activity_table.html
(function initJobDefinitions() {
  function lucideIcon(name) {
    return '<i data-lucide="' + name + '" class="w-6 h-6"></i>';
  }

  window.jobDefinitions = {
    trending_movies: {
      name: 'Trending Movies',
      description: 'Movies currently being watched and talked about on Trakt',
      icon: lucideIcon('trending-up'),
      iconColor: 'text-yellow-400',
      gradient: 'from-yellow-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.trending_movies'
    },
    popular_movies: {
      name: 'Popular Movies',
      description: 'Most watched movies on Trakt over various time periods',
      icon: lucideIcon('flame'),
      iconColor: 'text-red-400',
      gradient: 'from-red-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.popular_movies'
    },
    box_office: {
      name: 'Box Office',
      description: 'Top movies from the weekend box office',
      icon: lucideIcon('ticket'),
      iconColor: 'text-green-400',
      gradient: 'from-green-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.box_office'
    },
    anticipated_movies: {
      name: 'Anticipated Movies',
      description: 'Most anticipated upcoming movies',
      icon: lucideIcon('clock-3'),
      iconColor: 'text-purple-400',
      gradient: 'from-purple-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.anticipated_movies'
    },
    favorited_movies: {
      name: 'Favorited Movies',
      description: 'Most favorited movies on Trakt',
      icon: lucideIcon('heart'),
      iconColor: 'text-pink-400',
      gradient: 'from-pink-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.favorited_movies'
    },
    played_movies: {
      name: 'Played Movies',
      description: 'Most played movies over a time period',
      icon: lucideIcon('play'),
      iconColor: 'text-blue-400',
      gradient: 'from-blue-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.played_movies'
    },
    watched_movies: {
      name: 'Watched Movies',
      description: 'Most watched movies over a time period',
      icon: lucideIcon('eye'),
      iconColor: 'text-cyan-400',
      gradient: 'from-cyan-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.watched_movies'
    },
    collected_movies: {
      name: 'Collected Movies',
      description: 'Most collected movies over a time period',
      icon: lucideIcon('archive'),
      iconColor: 'text-amber-400',
      gradient: 'from-amber-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.collected_movies'
    },
    smart_popular_movies: {
      name: 'Smart Popular Movies',
      description: 'Adaptive popular movies with dynamic rating thresholds',
      icon: lucideIcon('brain'),
      iconColor: 'text-indigo-400',
      gradient: 'from-indigo-900/50 to-slate-800/50',
      category: 'Movies',
      configKey: 'jobs.smart_popular_movies'
    },
    trending_shows: {
      name: 'Trending Shows',
      description: 'TV shows currently being watched and talked about on Trakt',
      icon: lucideIcon('trending-up'),
      iconColor: 'text-yellow-400',
      gradient: 'from-yellow-900/50 to-slate-800/50',
      category: 'TV Shows',
      configKey: 'jobs.trending_shows'
    },
    popular_shows: {
      name: 'Popular Shows',
      description: 'Most watched TV shows on Trakt over various time periods',
      icon: lucideIcon('flame'),
      iconColor: 'text-red-400',
      gradient: 'from-red-900/50 to-slate-800/50',
      category: 'TV Shows',
      configKey: 'jobs.popular_shows'
    },
    anticipated_shows: {
      name: 'Anticipated Shows',
      description: 'Most anticipated upcoming TV shows',
      icon: lucideIcon('clock-3'),
      iconColor: 'text-purple-400',
      gradient: 'from-purple-900/50 to-slate-800/50',
      category: 'TV Shows',
      configKey: 'jobs.anticipated_shows'
    },
    favorited_shows: {
      name: 'Favorited Shows',
      description: 'Most favorited TV shows on Trakt',
      icon: lucideIcon('heart'),
      iconColor: 'text-pink-400',
      gradient: 'from-pink-900/50 to-slate-800/50',
      category: 'TV Shows',
      configKey: 'jobs.favorited_shows'
    },
    played_shows: {
      name: 'Played Shows',
      description: 'Most played TV shows over a time period',
      icon: lucideIcon('play'),
      iconColor: 'text-blue-400',
      gradient: 'from-blue-900/50 to-slate-800/50',
      category: 'TV Shows',
      configKey: 'jobs.played_shows'
    },
    watched_shows: {
      name: 'Watched Shows',
      description: 'Most watched TV shows over a time period',
      icon: lucideIcon('eye'),
      iconColor: 'text-cyan-400',
      gradient: 'from-cyan-900/50 to-slate-800/50',
      category: 'TV Shows',
      configKey: 'jobs.watched_shows'
    },
    collected_shows: {
      name: 'Collected Shows',
      description: 'Most collected TV shows over a time period',
      icon: lucideIcon('archive'),
      iconColor: 'text-amber-400',
      gradient: 'from-amber-900/50 to-slate-800/50',
      category: 'TV Shows',
      configKey: 'jobs.collected_shows'
    },
    smart_popular_shows: {
      name: 'Smart Popular Shows',
      description: 'Adaptive popular TV shows with dynamic rating thresholds',
      icon: lucideIcon('brain'),
      iconColor: 'text-indigo-400',
      gradient: 'from-indigo-900/50 to-slate-800/50',
      category: 'TV Shows',
      configKey: 'jobs.smart_popular_shows'
    }
  };
})();
