import axios from 'axios';

const baseURL = "https://95.165.132.221:67/api";

const api = axios.create({
    baseURL: baseURL, 
});

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    console.log(token);
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export {api, baseURL};
