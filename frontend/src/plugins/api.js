import axios from 'axios';

// Create axios instance with base configuration
const api = axios.create({
    baseURL: 'http://localhost:8080/api',
    headers: {
        'Content-Type': 'application/json'
    }
});

// Add response interceptor for error handling
api.interceptors.response.use(
    response => response,
    error => {
        console.error('API Error:', error);
        return Promise.reject(error);
    }
);

// Create Vue plugin
export default {
    install: (app) => {
        app.config.globalProperties.$api = api;
    }
};

// Export the api instance directly for composition API usage
export { api }; 