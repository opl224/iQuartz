import { useEffect, useState } from "react";
import { getProducts } from "../api/products";
import type { Product } from "../types";

const SummaryCards = () => {
  const [stats, setStats] = useState({
    total: 0,
    active: 0,
    lowStock: 0,
    value: 0,
  });

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const res = await getProducts("", 1, 1000);
        const products: Product[] = res.data.data;

        setStats({
          total: products.length,
          active: products.filter((p: Product) => p.is_active).length,
          lowStock: products.filter((p: Product) => p.stock <= 10).length,
          value: products.reduce(
            (acc: number, curr: Product) => acc + curr.price * curr.stock,
            0,
          ),
        });
      } catch (err) {
        console.error(err);
      }
    };
    fetchStats();
  }, []);

  const formatRupiah = (num: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(num);

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(4, 1fr)",
        gap: "1rem",
        marginBottom: "2rem",
      }}
    >
      {[
        { label: "Total Produk", val: stats.total },
        { label: "Produk Aktif", val: stats.active },
        { label: "Stok Menipis (≤10)", val: stats.lowStock },
        { label: "Nilai Aset Stok", val: formatRupiah(stats.value) },
      ].map((item, idx) => (
        <div key={idx} className="summary-card">
          <div
            style={{
              fontSize: "0.8rem",
              color: "#6b7280",
              marginBottom: "0.5rem",
            }}
          >
            {item.label}
          </div>
          <div
            style={{ fontSize: "1.5rem", fontWeight: "bold", color: "#111827" }}
          >
            {item.val}
          </div>
        </div>
      ))}
    </div>
  );
};

export default SummaryCards;
