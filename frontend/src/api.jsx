import axios from 'axios';

const baseURL = "https://95.165.132.221:67/api";
const wsURL = "wss://95.165.132.221:67/api"

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


api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      console.log('hola');
      const refreshToken = localStorage.getItem('refresh_token');
      const loginResponse = await fetch(baseURL + '/refresh', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: refreshToken
        }),
      });

      const data = await loginResponse.json();

      localStorage.setItem('token', data.jwt_token);
      return api(originalRequest);
    }
  }
);

export {api, baseURL, wsURL};
