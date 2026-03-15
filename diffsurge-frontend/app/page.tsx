import { Header } from "@/components/landing/header";
import { Hero } from "@/components/landing/hero";
import { Problem } from "@/components/landing/problem";
import { HowItWorks } from "@/components/landing/how-it-works";
import { Features } from "@/components/landing/features";
import { CodeDemo } from "@/components/landing/code-demo";
import { OpenSource } from "@/components/landing/open-source";
import { Comparison } from "@/components/landing/comparison";
import { UseCases } from "@/components/landing/use-cases";
import { CTA } from "@/components/landing/cta";
import { Footer } from "@/components/landing/footer";
import { JsonLd } from "@/components/landing/json-ld";

export default function Home() {
  return (
    <>
      <JsonLd />
      <Header />
      <main>
        <Hero />
        <Problem />
        <HowItWorks />
        <Features />
        <CodeDemo />
        <OpenSource />
        <Comparison />
        <UseCases />
        <CTA />
      </main>
      <Footer />
    </>
  );
}
