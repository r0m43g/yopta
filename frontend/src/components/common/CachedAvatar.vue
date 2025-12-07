<template>
  <div class="flex items-center justify-center">
    <!-- Loading placeholder that appears until the avatar is fully loaded -->
    <div v-if="!isLoaded && !imageError" class="avatar-placeholder">
      <div class="loading loading-spinner loading-xs"></div>
    </div>

    <!-- The main avatar image - hidden until loaded, shown when ready -->
    <img
      v-show="isLoaded && !imageError"
      :src="cachedSrc"
      :alt="alt"
      @load="handleImageLoaded"
      @error="handleImageError"
    />

    <!-- Fallback image shown when the main avatar fails to load -->
    <img
      v-if="imageError"
      :src="defaultSrc"
      :alt="alt"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';

/**
 * The Global Image Memory Bank - where avatars go to avoid being forgotten!
 *
 * This singleton cache stores images we've already loaded to avoid unnecessary
 * network requests. Think of it as a photo album that remembers faces so you
 * don't have to keep asking "who are you again?" every time you see someone.
 */
const imageCache = {
  // A map of promises that represent ongoing image loading operations
  promises: new Map(),

  // A map of statuses to track the loading state of each image
  statuses: new Map(),

  /**
   * The Time Traveler - loads images before you even know you need them!
   *
   * Preloads an image by starting the download process early, so it's ready
   * when you actually want to display it. Like preparing breakfast the night before,
   * this saves time when you're actually hungry in the morning.
   *
   * @param {string} src - Where to find the image (URL)
   * @param {string} defaultSrc - Where to find a backup image if the main one fails
   * @returns {Promise} - A promise that resolves when the image is loaded
   */
  preload(src, defaultSrc) {
    // If already loading or loaded, return existing promise
    if (this.promises.has(src)) {
      return this.promises.get(src);
    }

    // If no avatar is set, use default
    if (!src || src === 'none') {
      return Promise.resolve(defaultSrc);
    }

    // Create a new promise for loading this image
    const promise = new Promise((resolve, reject) => {
      const img = new Image();

      img.onload = () => {
        // Successfully loaded - mark as complete and resolve promise
        this.statuses.set(src, 'loaded');
        resolve(src);
      };

      img.onerror = () => {
        // Failed to load - mark as error and resolve with default image
        this.statuses.set(src, 'error');
        resolve(defaultSrc);
      };

      // Start loading
      img.src = src;
    });

    // Store promise in cache
    this.promises.set(src, promise);
    return promise;
  },

  /**
   * The Status Checker - reports on an image's loading progress!
   *
   * Returns the current loading status of an image.
   * Like a construction foreman checking if a building is finished,
   * this function reports whether an image is still in progress,
   * successfully completed, or failed catastrophically.
   *
   * @param {string} src - The URL of the image to check
   * @returns {string} - The loading status ('pending', 'loaded', or 'error')
   */
  getStatus(src) {
    return this.statuses.get(src) || 'pending';
  },

  /**
   * The Selective Forgetter - removes specific images from memory!
   *
   * Clears a specific image from the cache to force reloading.
   * Like deliberately forgetting your ex's birthday, this function
   * removes a particular image from memory so you can start fresh.
   *
   * @param {string} src - The URL of the image to forget
   */
  clear(src) {
    this.promises.delete(src);
    this.statuses.delete(src);
  },

  /**
   * The Memory Wiper - completely erases the image database!
   *
   * Clears the entire cache, like a digital lobotomy.
   * For when you want to start with a completely blank slate,
   * this function purges all image knowledge like it never existed.
   */
  clearAll() {
    this.promises.clear();
    this.statuses.clear();
  }
};

/**
 * Component Props - the customization knobs for our avatar display!
 *
 * These define how the avatar should look and behave, like instructions
 * for a portrait artist on how to paint your likeness.
 */
const props = defineProps({
  // The path to the avatar image
  src: {
    type: String,
    default: null
  },
  // Alternative text for accessibility
  alt: {
    type: String,
    default: 'Naudotojo avataras'
  },
  // Path to default image to use if avatar fails
  defaultSrc: {
    type: String,
    default: '/avatars/yopta.webp'
  },
  // Size of the avatar
  size: {
    type: [String, Number],
    default: '100%'
  },
  // Whether the avatar should be rendered as a circle
  rounded: {
    type: Boolean,
    default: false
  }
});

/**
 * Component State - the internal memory of our avatar component!
 *
 * These reactive references track the loading state and appearance of our avatar,
 * like a portrait artist keeping notes about their progress on your painting.
 */
const isLoaded = ref(false);
const imageError = ref(false);
const cachedSrc = ref(null);
const isPreloading = ref(false);

/**
 * The Path Formatter - determines where to find the avatar image!
 *
 * Computes the proper path to the avatar image based on environment and props.
 * Like a detective piecing together clues to find a person's whereabouts,
 * this function figures out the exact location of your avatar image.
 */
const formattedSrc = computed(() => {
  // Determine operating mode
  const isDev = import.meta.env.DEV;
  const avatarUrl = import.meta.env.VITE_AVATAR_BASE_URL || '';
  const src = (!props.src || props.src === 'none') ? props.defaultSrc : props.src;

  // For development mode, use full URL to backend
  if (isDev && avatarUrl) {
    const cleanPath = src.startsWith('/') ? src.substring(1) : src;
    return `${avatarUrl}/${cleanPath}`;
  }

  // If it's already a full URL
  if (props.src && props.src.startsWith('http')) {
    return props.src;
  }

  // If path already starts with slash
  if (props.src && props.src.startsWith('/')) {
    return props.src;
  }

  // Otherwise add slash
  return props.src ? '/' + props.src : props.defaultSrc;
});

/**
 * The Avatar Detective - reacts when the source changes to find a new image!
 *
 * Watches for changes to the src prop and reloads the avatar when needed.
 * Like a vigilant detective who notices when a person changes their appearance,
 * this watcher springs into action whenever your avatar source changes.
 */
watch(() => props.src, (newSrc) => {
  preloadImage();
}, { immediate: false });

/**
 * The Image Preloader - prepares your avatar before you need it!
 *
 * Loads the image in advance and manages loading states.
 * Like a diligent assistant who brings you coffee before you ask,
 * this function ensures your avatar is ready to display as soon as possible.
 */
function preloadImage() {
  if (isPreloading.value) return;

  isPreloading.value = true;
  isLoaded.value = false;
  imageError.value = false;

  const formattedSrcValue = formattedSrc.value;

  // Check image loading status
  const status = imageCache.getStatus(formattedSrcValue);

  if (status === 'loaded') {
    // If image already loaded, just set it
    cachedSrc.value = formattedSrcValue;
    isLoaded.value = true;
    isPreloading.value = false;
  }
  else if (status === 'error') {
    // If previous loading failed, show error state
    imageError.value = true;
    isPreloading.value = false;
  }
  else {
    // Start loading image
    imageCache.preload(formattedSrcValue, props.defaultSrc)
      .then(src => {
        cachedSrc.value = src;
        if (src === props.defaultSrc && formattedSrcValue !== props.defaultSrc) {
          imageError.value = true;
        } else {
          isLoaded.value = true;
        }
      })
      .finally(() => {
        isPreloading.value = false;
      });
  }
}

/**
 * The Success Celebrator - acknowledges when an image loads successfully!
 *
 * Handles the successful loading of an image by updating component state.
 * Like a proud parent celebrating when their child finally learns to ride a bike,
 * this function updates our component state when an image loads correctly.
 */
function handleImageLoaded() {
  isLoaded.value = true;
  imageError.value = false;
}

/**
 * The Failure Manager - gracefully handles image loading disasters!
 *
 * Processes error events when images fail to load.
 * Like a calm crisis manager who implements the backup plan when things go wrong,
 * this function ensures your users still see something even when the main image fails.
 */
function handleImageError() {
  imageError.value = true;
  isLoaded.value = false;
  imageCache.clear(formattedSrc.value);
}
</script>

<style scoped>
.avatar-container {
  position: relative;
  overflow: hidden;
  display: inline-block;
}

.avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.1);
}

.avatar-image {
  display: block;
  transition: opacity 0.3s ease;
}

.avatar-loaded .avatar-image {
  opacity: 1;
}

.fallback {
  opacity: 1;
}
</style>
