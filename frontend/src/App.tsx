import SummaryCards from "./components/SummaryCards";
import ProductTable from "./components/ProductTable";
import OrderForm from "./components/OrderForm";

function App() {
  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <header className="mb-8">
        <h1 className="text-3xl font-bold text-gray-800">StockFlow ERP</h1>
        <p className="text-gray-600">Mini Inventory & Sales Module</p>
      </header>

      <main>
        {/* 1. Summary Cards */}
        <SummaryCards />

        {/* 2. Product Table */}
        <ProductTable />

        {/* 3. Order Form */}
        <OrderForm />
      </main>
    </div>
  );
}

export default App;
