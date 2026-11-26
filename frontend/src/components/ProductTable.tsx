import { useEffect, useState } from "react";
import { getProducts } from "../api/products";
import type { Product } from "../types";

const ProductTable = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<"all" | "active" | "inactive">("all");

  // Fungsi untuk format Rupiah
  const formatRupiah = (price: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(price);
  };

  // Fungsi untuk waktu relatif sederhana
  const timeAgo = (dateString: string | null) => {
    if (!dateString) return "Never";
    const date = new Date(dateString);
    const now = new Date();
    const diffInHours = Math.floor(
      (now.getTime() - date.getTime()) / (1000 * 60 * 60),
    );
    if (diffInHours < 1) return "Just now";
    return `${diffInHours}h ago`;
  };

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await getProducts(
        search === "" ? undefined : search,
        1,
        100,
      ); // Ambil semua untuk filter lokal dulu
      let data = response.data.data;

      // Filter lokal untuk Active/Inactive
      if (filter === "active") data = data.filter((p) => p.is_active);
      if (filter === "inactive") data = data.filter((p) => !p.is_active);

      setProducts(data);
    } catch (err) {
      setError("Gagal memuat data. Pastikan backend berjalan.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [search, filter]); // Re-fetch saat search atau filter berubah

  if (loading) return <div className="p-4">Loading products...</div>;
  if (error)
    return (
      <div className="p-4 text-red-500">
        {error}{" "}
        <button onClick={fetchData} className="underline">
          Retry
        </button>
      </div>
    );

  return (
    <div className="p-4">
      <div className="flex justify-between mb-4">
        <input
          type="text"
          placeholder="Cari SKU atau Nama..."
          className="border p-2 rounded"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select
          className="border p-2 rounded"
          value={filter}
          onChange={(e) => setFilter(e.target.value as any)}
        >
          <option value="all">Semua Status</option>
          <option value="active">Aktif</option>
          <option value="inactive">Nonaktif</option>
        </select>
      </div>

      <table className="w-full border-collapse border">
        <thead>
          <tr className="bg-gray-100">
            <th className="border p-2">SKU</th>
            <th className="border p-2">Nama</th>
            <th className="border p-2">Harga</th>
            <th className="border p-2">Stok</th>
            <th className="border p-2">Status</th>
            <th className="border p-2">Gerakan Terakhir</th>
            <th className="border p-2">Aksi</th>
          </tr>
        </thead>
        <tbody>
          {products.length === 0 ? (
            <tr>
              <td colSpan={7} className="text-center p-4">
                Tidak ada produk ditemukan.
              </td>
            </tr>
          ) : (
            products.map((p) => (
              <tr key={p.id} className="hover:bg-gray-50">
                <td className="border p-2 font-mono">{p.sku}</td>
                <td className="border p-2">{p.name}</td>
                <td className="border p-2 text-right">
                  {formatRupiah(p.price)}
                </td>
                <td className="border p-2 text-center">{p.stock}</td>
                <td className="border p-2 text-center">
                  <span
                    className={`px-2 py-1 rounded text-xs ${p.is_active ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"}`}
                  >
                    {p.is_active ? "Active" : "Inactive"}
                  </span>
                </td>
                <td className="border p-2 text-center text-sm text-gray-500">
                  {timeAgo(p.last_movement_at)}
                </td>
                <td className="border p-2 text-center">
                  <button className="text-blue-600 hover:underline mr-2">
                    Edit
                  </button>
                  <button className="text-blue-600 hover:underline">
                    Stock
                  </button>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
};

export default ProductTable;
