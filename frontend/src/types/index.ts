export interface Product {
  id: number;
  sku: string;
  name: string;
  price: number;
  stock: number;
  is_active: boolean;
  created_at: string;
  last_movement_at?: string | null;
}

export interface ApiResponse<T> {
  data: T;
  page?: number;
  limit?: number;
  total?: number;
}

export interface OrderItem {
  sku: string;
  qty: number;
  unit_price: number;
}

export interface Order {
  id: number;
  customer: string;
  status: "CREATED" | "CANCELLED";
  total: number;
  items: OrderItem[];
  created_at: string;
  cancelled_at?: string | null;
}
