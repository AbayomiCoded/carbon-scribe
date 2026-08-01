import {
  Controller,
  Get,
  Post,
  Put,
  Delete,
  Body,
  Param,
  Query,
  UseGuards,
} from '@nestjs/common';
import { SearchService } from './services/search.service';
import { RecommendationService } from './services/recommendation.service';
import { DiscoveryService } from './services/discovery.service';
import { StatsService } from './services/stats.service';
import { SearchQueryDto } from './dto/search-query.dto';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { JwtPayload } from '../auth/interfaces/jwt-payload.interface';
import { MarketplaceService } from './marketplace.service';
import { RateLimit, RateLimits } from '../rate-limit/rate-limit.decorator';

@UseGuards(JwtAuthGuard)
@Controller('api/v1/marketplace')
export class MarketplaceController {
  constructor(
    private readonly searchService: SearchService,
    private readonly recommendationService: RecommendationService,
    private readonly discoveryService: DiscoveryService,
    private readonly statsService: StatsService,
    private readonly marketplaceService: MarketplaceService,
  ) {}

  /**
   * Search credits
   * Rate limited to 20 requests per minute per IP
   */
  @Get('search')
  @RateLimit(RateLimits.SEARCH)
  async search(@Query() query: SearchQueryDto) {
    return this.searchService.search(query);
  }

  @Get('recommendations')
  async recommendations(
    @CurrentUser() user: JwtPayload,
    @Query('sdgs') sdgsParam?: string,
  ) {
    const sdgs =
      sdgsParam && sdgsParam.length > 0
        ? sdgsParam
            .split(',')
            .map((p) => Number(p.trim()))
            .filter((n) => !Number.isNaN(n))
        : undefined;

    return this.recommendationService.getRecommendations(
      {
        userId: user.sub,
        companyId: user.companyId,
      },
      { sdgs },
    );
  }

  @Get('featured')
  async featured() {
    return this.discoveryService.getFeatured();
  }

  @Get('trending')
  async trending() {
    return this.discoveryService.getTrending();
  }

  @Get('new')
  async newest() {
    return this.discoveryService.getNewest();
  }

  @Get('stats')
  async stats() {
    return this.statsService.getStats();
  }

  @Get('filters')
  async filters() {
    return this.statsService.getFilters();
  }

  /**
   * Get similar credits
   * Rate limited to 30 requests per minute per IP
   */
  @Get('similar/:creditId')
  @RateLimit(RateLimits.SIMILAR_CREDITS)
  async similar(@Param('creditId') creditId: string) {
    return this.marketplaceService.getSimilarCredits(creditId);
  }

  @Get('discovery')
  async discoveryOverview() {
    return this.discoveryService.getDiscoveryOverview();
  }

  // ── Cart endpoints ──────────────────────────────────────────────────────────

  @Get('credits')
  async getCredits(@Query() query: any) {
    return this.marketplaceService.getCredits(query);
  }

  @Get('credits/:id')
  async getCreditById(@Param('id') id: string) {
    return this.marketplaceService.getCreditById(id);
  }

  @Post('cart')
  async addToCart(
    @CurrentUser() user: JwtPayload,
    @Body() dto: any,
  ) {
    return this.marketplaceService.addToCart(user.companyId, user.sub, dto);
  }

  @Get('cart')
  async getCart(@CurrentUser() user: JwtPayload) {
    return this.marketplaceService.getCart(user.companyId, user.sub);
  }

  @Delete('cart/:itemId')
  async removeFromCart(
    @CurrentUser() user: JwtPayload,
    @Param('itemId') itemId: string,
  ) {
    return this.marketplaceService.removeFromCart(user.companyId, user.sub, itemId);
  }

  @Post('checkout')
  async checkout(
    @CurrentUser() user: JwtPayload,
    @Body() dto: any,
  ) {
    return this.marketplaceService.checkout(user.companyId, user.sub, dto);
  }
}