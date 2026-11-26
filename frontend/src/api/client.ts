import axios from "axios";

const api = axios.create({
  baseURL: "/api", // Ganti URL lengkap menjadi relatif
});

export default api;
