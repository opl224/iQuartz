import api from "./client";
import type { Product, ApiResponse } from "../types"; // Tambahkan 'type' di sini

export const getProducts = (
  search?: string,
  page: number = 1,
  limit: number = 10,
) => {
  return api.get<ApiResponse<Product[]>>("/products", {
    params: { search, page, limit },
  });
};
