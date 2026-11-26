import api from "./client";
import type { Order } from "../types";

export const createOrder = (data: {
  customer: string;
  items: { sku: string; qty: number }[];
}) => {
  return api.post("/orders", data);
};

export const getOrder = (id: number) => {
  return api.get<Order>(`/orders/${id}`);
};

export const cancelOrder = (id: number) => {
  return api.post(`/orders/${id}/cancel`);
};
