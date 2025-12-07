// stores/auth.js
import { defineStore } from 'pinia'
import { jwtDecode } from 'jwt-decode'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    refreshToken: localStorage.getItem('refreshToken') || null,
    accessToken: null,
    csrfToken: null,
    user: null, // user id
    username: null,
    email: null,
    role: null,
    theme: null,
    avatar: null,
  }),
  actions: {
    /**
     * The JWT token whisperer - converts cryptic strings into user identities!
     * Sets the access token and reveals the hidden secrets (user claims) inside it.
     * Like an archaeologist carefully extracting artifacts, this function decodes
     * the JWT to find the treasure of user information within.
     *
     * @param {string} token - The magical JWT string that grants access to the kingdom
     */
    setAccessToken(token) {
      this.accessToken = token
      // Save token to localStorage for persistent sessions
      localStorage.setItem('accessToken', token)

      try {
        const decoded = jwtDecode(token)
        this.user = decoded
      } catch (error) {
        this.user = null
        console.error('JWT dekodavimo klaida:', error)
      }
    },

    /**
     * The token maintenance specialist - ensures you can get a new access pass when needed!
     * Stores the refresh token that will be used to obtain new access tokens when they expire.
     * Think of it as keeping a spare key safely tucked away for when your main key stops working.
     *
     * @param {string} token - The refresh token - your "get out of 401 errors free" card
     */
    setRefreshToken(token) {
      this.refreshToken = token
      // Save to localStorage for session recovery
      localStorage.setItem('refreshToken', token)
    },

    /**
     * The security guard that protects against CSRF attacks like a bouncer at an exclusive club.
     * Sets the CSRF token that helps verify that requests are coming from your application
     * and not from some malicious site trying to impersonate your users.
     *
     * @param {string} csrfToken - The token that proves "It's really me making this request!"
     */
    setCsrfToken(csrfToken) {
      this.csrfToken = csrfToken
    },

    /**
     * The grand eraser - wipes all traces of your identity from the system!
     * Clears all authentication and user data from the store and localStorage.
     * Like a digital cleaning service that removes all evidence you were ever logged in.
     * Use during logout or when the session has gone stale like week-old bread.
     */
    clearToken() {
      this.accessToken = null
      this.refreshToken = null
      this.user = null
      this.username = null
      this.email = null
      this.role = null
      this.theme = null
      this.avatar = null

      // Clean localStorage like a professional tidier
      localStorage.removeItem('refreshToken')
      localStorage.removeItem('accessToken')
    },

    /**
     * The identity establisher - properly sets up who you are in the system!
     * Populates the store with user profile information retrieved from the server.
     * Like putting on a name tag at a conference, this makes sure the app knows exactly who you are.
     *
     * @param {Object} user - The user object containing all your personal details and preferences
     */
    setUser(user) {
      this.username = user.username
      this.email = user.email
      this.role = user.role
      this.theme = user.theme
      this.avatar = user.avatar === 'none' ? null : user.avatar
    },

    /**
     * The face swapper - updates your digital appearance!
     * Updates the user's avatar path in the store. Whether you're going for a professional
     * headshot or that picture of you with your cat, this function makes it happen.
     *
     * @param {string} avatarPath - The path to your new digital face, or null to go incognito
     */
    setAvatar(avatarPath) {
      this.avatar = avatarPath
    }
  },
  getters: {
    /**
     * The authentication detective - investigates if you're really who you claim to be!
     * Checks if the user has a valid, non-expired JWT token which is the digital equivalent
     * of having an up-to-date ID card. Without this, you're just a digital ghost to the system.
     *
     * @returns {boolean} - True if you're legit, false if you're a digital nobody
     */
    isAuthenticated(state) {
      if (!state.accessToken) return false

      try {
        const decoded = jwtDecode(state.accessToken)
        return decoded.exp * 1000 > Date.now()
      } catch (error) {
        return false
      }
    },

    /**
     * The privilege inspector - determines if you have the keys to the kingdom!
     * Checks if the current user has admin privileges, which is like having an all-access
     * backstage pass at a concert. Regular users get the standard experience, but admins
     * get to see how the magic happens.
     *
     * @returns {boolean} - True if you're basically a digital god, false if you're a mere mortal
     */
    isAdmin(state) {
      return state.role === 'admin'
    },

    /**
     * The aesthetic advisor - tells you what visual style you prefer!
     * Returns the user's theme preference, which controls the app's appearance.
     * Whether you're a dark mode vampire or a light mode sunshine lover, this
     * function knows your visual comfort zone.
     *
     * @returns {string} - The name of your chosen theme, your digital interior decorator
     */
    getTheme(state) {
      return state.theme
    },

    /**
     * The portrait provider - fetches your digital face!
     * Returns the path to the user's avatar image, which is like your digital passport photo.
     * The way other users recognize you in the system, unless you're going incognito.
     *
     * @returns {string|null} - Path to your digital face, or null if you're in stealth mode
     */
    getAvatar(state) {
      return state.avatar
    },

    /**
     * The identity number fetcher - gives you your unique digital serial number!
     * Returns the user's ID, which is like your social security number in the system.
     * This unique identifier is how the database knows it's talking about YOU and not
     * some other user with a suspiciously similar username.
     *
     * @returns {number|null} - Your unique ID in the system, or null if you don't exist (spooky!)
     */
    getID(state) {
      return state.user?.user_id || null
    },
  },
})
